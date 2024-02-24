package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Fserlut/gophermart/internal/handlers"
)

func NewRouter(handler *handlers.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Compress(5, "application/json", "text/html"))

	router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", handler.Register)
			r.Post("/login", handler.Login)
		})
	})

	return router
}
