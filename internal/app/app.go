package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/db"
	"github.com/Fserlut/gophermart/internal/handlers"
	"github.com/Fserlut/gophermart/internal/router"
	"github.com/Fserlut/gophermart/internal/services"
)

type App struct {
	Server *http.Server
	Router *chi.Mux
	logger *slog.Logger
	config *config.Config
}

func CreateApp(logger *slog.Logger, cfg *config.Config) *App {
	//TODO норм ли это создавать все именно тут?
	userRepository, err := db.NewDB(cfg.DatabaseURI)
	if err != nil {
		logger.Error("error on init db: ", err.Error())
		panic(err.Error())
	}

	userService := services.NewUserService(userRepository)

	handler := handlers.NewHandler(logger, userService)

	r := router.NewRouter(handler)

	return &App{
		Router: r,
		logger: logger,
		config: cfg,
		Server: &http.Server{
			Addr:    cfg.RunAddress,
			Handler: r,
		},
	}
}

func (a *App) Run() {
	a.logger.Info(fmt.Sprintf("Server starting on %s", a.config.RunAddress))

	if err := a.Server.ListenAndServe(); err != nil {
		panic(err)
	}
}
