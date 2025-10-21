package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// requests
	PlaceBid MessageKind = iota

	// success
	SuccessfullyPlacedBid

	// errors
	FailedToPlaceBid
	InvalidJson

	// info
	NewBidPlaced
	AuctionEnded
)

type Message struct {
	UserId    uuid.UUID   `json:"user_id,omitempty"`
	Message   string      `json:"message,omitempty"`
	Kind      MessageKind `json:"kind"`
	BidAmount float64     `json:"bid_amount,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsService BidsService
}

func (r *AuctionRoom) registerClient(client *Client) {
	slog.Info("New user connected", "Client", client)
	r.Clients[client.UserID] = client
}

func (r *AuctionRoom) unregisterClient(client *Client) {
	slog.Info("User disconnected", "Client", client)
	delete(r.Clients, client.UserID)
}

func (r *AuctionRoom) broadcastMessage(message Message) {
	slog.Info("Broadcasting message", "Room", r.Id, "Message", message, "UserId", message.UserId)
	switch message.Kind {
	case PlaceBid:
		bid, err := r.BidsService.PlaceBid(r.Context, r.Id, message.UserId, message.BidAmount)
		if err != nil {
			if errors.Is(err, ErrBidAmountTooLow) {
				if client, ok := r.Clients[message.UserId]; ok {
					client.Send <- Message{
						Message: ErrBidAmountTooLow.Error(),
						Kind:    FailedToPlaceBid,
						UserId:  message.UserId,
					}
					return
				}
			}
			if client, ok := r.Clients[message.UserId]; ok {
				client.Send <- Message{
					Message: err.Error(),
					Kind:    FailedToPlaceBid,
					UserId:  message.UserId,
				}
			}
		}
		if client, ok := r.Clients[message.UserId]; ok {
			client.Send <- Message{
				Message: fmt.Sprintf("Your bid of %.2f was placed successfully", bid.BidAmount),
				Kind:    SuccessfullyPlacedBid,
				UserId:  message.UserId,
			}
		}
		for id, client := range r.Clients {
			newBidMessage := Message{
				UserId:    message.UserId,
				Message:   fmt.Sprintf("New bid of %.2f was placed by %s", bid.BidAmount, id),
				Kind:      NewBidPlaced,
				BidAmount: bid.BidAmount,
			}
			if id == message.UserId {
				continue
			}
			client.Send <- newBidMessage
		}
	case InvalidJson:
		client, ok := r.Clients[message.UserId]
		if !ok {
			slog.Info("User not found", "UserId", message.UserId)
			return
		}
		client.Send <- Message{
			Message: "Invalid JSON",
			Kind:    InvalidJson,
			UserId:  message.UserId,
		}
	}
}

func (r *AuctionRoom) Run() {
	slog.Info("Auction has started", "Room:", r.Id)
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended", "Room", r.Id)
			for _, client := range r.Clients {
				client.Send <- Message{
					Message: "Auction has ended",
					Kind:    AuctionEnded,
				}
				return
			}
		}
	}
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Context:     ctx,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		BidsService: bidsService,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserID uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserID: userID,
	}
}

const (
	maxMessageSize = 512
	readDeadLine   = 60 * time.Second
	pingPeriod     = (readDeadLine * 9) / 10
	writeWait      = 10 * time.Second
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))
		return nil
	})
	for {
		var m Message
		m.UserId = c.UserID
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("unexpected close error", "Error", err)
				return
			}
			c.Room.Broadcast <- Message{
				Message: "Invalid JSON",
				Kind:    InvalidJson,
				UserId:  m.UserId,
			}
			continue
		}
		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Message: "Connection closed",
					Kind:    websocket.CloseMessage,
				})
				return
			}
			if message.Kind == AuctionEnded {
				close(c.Send)
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Room.Unregister <- c
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("failed to write ping message", "Error", err)
				return
			}
		}

	}
}
