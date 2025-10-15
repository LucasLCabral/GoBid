package api

import (
	"net/http"

	"github.com/LucasLCabral/go-bid/internal/jsonutils"
	"github.com/gorilla/csrf"
)

func (a *API) HandleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"csrf_token": token,
	})
}


func (a *API) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Sessions.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"error": "must be logged in",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
