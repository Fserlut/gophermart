package app

import (
	"fmt"
	order2 "github.com/Fserlut/gophermart/internal/handlers/order"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/db"
	user2 "github.com/Fserlut/gophermart/internal/handlers/user"
	"github.com/Fserlut/gophermart/internal/router"
	"github.com/Fserlut/gophermart/internal/services/order"
	"github.com/Fserlut/gophermart/internal/services/user"
)

type App struct {
	Server *http.Server
	Router *chi.Mux
	logger *slog.Logger
	config *config.Config
}

func CreateApp(logger *slog.Logger, cfg *config.Config) (*App, error) {
	userRepository, err := db.NewDB(cfg.DatabaseURI)
	if err != nil {
		logger.Error("error on init db: ", err.Error())
		return nil, err
	}

	userService := user.NewUserService(userRepository)
	orderService := order.NewOrderService(logger, userRepository, cfg)

	userHandler := user2.NewUserHandler(logger, userService)

	orderHandler := order2.NewOrderHandler(logger, orderService)

	r := router.NewRouter(userHandler, orderHandler)

	return &App{
		Router: r,
		logger: logger,
		config: cfg,
		Server: &http.Server{
			Addr:    cfg.RunAddress,
			Handler: r,
		},
	}, nil
}

func (a *App) Run() {
	a.logger.Info(fmt.Sprintf("Server starting on %s", a.config.RunAddress))

	if err := a.Server.ListenAndServe(); err != nil {
		panic(err)
	}
}
