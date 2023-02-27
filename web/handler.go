package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewHandler() *Handler {
	handler := &Handler{
		chi.NewRouter(),
	}

	handler.Use(middleware.Logger)

	handler.Get("/", handler.WebShowShops())
	handler.Get("/create-shop", handler.WebCreateShopForm())
	handler.Get("/connexion", handler.WebCreateConnexionForm())
	/*handler.Get("/create-account", handler.WebCreateAccountForm())*/

	handler.Route("/api", func(r chi.Router) {
		//r.Post("/", handler.AddCreateAccount())
	})

	return handler
}

type Handler struct {
	*chi.Mux
}
