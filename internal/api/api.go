package api

import (
	"github.com/LucasLCabral/go-bid/internal/services"
	"github.com/gorilla/websocket"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type API struct {
	Router          *chi.Mux
	UserService     *services.UserService
	Sessions        *scs.SessionManager
	ProductsService *services.ProductsService
	WSUpgrader      *websocket.Upgrader
	AuctionLobby    *services.AuctionLobby
	BidsService     *services.BidsService
}
