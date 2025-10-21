package api

import (
	"errors"
	"net/http"

	"github.com/LucasLCabral/go-bid/internal/jsonutils"
	"github.com/LucasLCabral/go-bid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *API) HandleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductID := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(rawProductID)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error": "invalid product id",
		})
		return
	}
	_, err = a.ProductsService.GetProductByID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"error": "product not found",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}
	userId, ok := a.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": "unexpected error",
		})
		return
	}

	conn, err := a.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to upgrade to websocket: " + err.Error(),
		})
		return
	}

	a.AuctionLobby.Lock()
	room, ok := a.AuctionLobby.Rooms[productID]
	a.AuctionLobby.Unlock()
	if !ok {
		conn.WriteJSON(map[string]any{
			"error": "auction has ended",
		})
		conn.Close()
		return
	}

	client := services.NewClient(room, conn, userId)

	room.Register <- client
	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
