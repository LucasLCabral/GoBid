package api

import "github.com/go-chi/chi/v5"

func (a *API) BindRoutes() {
	a.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func (r chi.Router)  {
			r.Route("/users", func (r chi.Router)  {
				r.Post("/signup", a.HandleSignUpUser)
				r.Post("/Login", a.HandleLoginUser)
				r.Post("/Login", a.HandleLogoutUser)
			})
		})
	})
}
