package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	authMiddleware "github.com/Fserlut/gophermart/internal/handlers/middleware"
	"github.com/Fserlut/gophermart/internal/handlers/order"
	"github.com/Fserlut/gophermart/internal/handlers/user"
)

func NewRouter(userHandler *user.UserHandler, orderHandler *order.OrderHandler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Compress(5, "application/json", "text/html", "text/plain"))

	router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)

			r.With(authMiddleware.AuthMiddleware).Post("/orders", orderHandler.CreateOrder)
			r.With(authMiddleware.AuthMiddleware).Get("/orders", orderHandler.GetOrders)
			r.With(authMiddleware.AuthMiddleware).Get("/balance", orderHandler.GetUserBalance)
			r.With(authMiddleware.AuthMiddleware).Post("/balance/withdraw", orderHandler.Withdraw)
			r.With(authMiddleware.AuthMiddleware).Get("/withdrawals", orderHandler.Withdrawals)
		})
	})

	return router
}
