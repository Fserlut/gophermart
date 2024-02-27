package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Fserlut/gophermart/internal/handlers"
	authMiddleware "github.com/Fserlut/gophermart/internal/handlers/middleware"
)

func NewRouter(handler *handlers.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Compress(5, "application/json", "text/html", "text/plain"))

	router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", handler.Register)
			r.Post("/login", handler.Login)

			r.With(authMiddleware.AuthMiddleware).Post("/orders", handler.CreateOrder)
			r.With(authMiddleware.AuthMiddleware).Get("/orders", handler.GetOrders)
			r.With(authMiddleware.AuthMiddleware).Get("/balance", handler.GetUserBalance)
			r.With(authMiddleware.AuthMiddleware).Post("/balance/withdraw", handler.Withdraw)
			r.With(authMiddleware.AuthMiddleware).Get("/withdrawals", handler.Withdrawals)
		})
	})

	return router
}
