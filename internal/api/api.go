package api

import (
	"github.com/LucasLCabral/go-bid/internal/services"
	"github.com/go-chi/chi/v5"
)

type API struct {
	Router *chi.Mux
	UserService *services.UserService
}
