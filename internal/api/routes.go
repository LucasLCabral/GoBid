package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *API) BindRoutes() {
	a.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, a.Sessions.LoadAndSave)

	// csrfMiddleware := csrf.Protect(
	// 	[]byte(os.Getenv("GOBID_CSRF_KEY")),
	// 	csrf.Secure(false), // dev only
	// )
	// a.Router.Use(csrfMiddleware)

	a.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// r.Get("/csrftoken", a.HandleGetCSRFToken)
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.HandleSignUpUser)
				r.Post("/login", a.HandleLoginUser)
				r.With(a.AuthMiddleware).Post("/logout", a.HandleLogoutUser)
			})

			r.Route("/products", func(r chi.Router) {
				r.Post("/", a.HandleCreateProduct)
			})
		})
	})
}
