package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/db"
	order2 "github.com/Fserlut/gophermart/internal/handlers/order"
	user2 "github.com/Fserlut/gophermart/internal/handlers/user"
	"github.com/Fserlut/gophermart/internal/logger"
	"github.com/Fserlut/gophermart/internal/router"
	"github.com/Fserlut/gophermart/internal/services/accrual"
	"github.com/Fserlut/gophermart/internal/services/order"
	"github.com/Fserlut/gophermart/internal/services/user"
)

type App struct {
	Server         *http.Server
	Router         *chi.Mux
	logger         logger.Logger
	config         *config.Config
	accrualService *accrual.Accrual
}

func CreateApp(logger logger.Logger, cfg *config.Config) (*App, error) {
	userRepository, err := db.NewDB(cfg.DatabaseURI)
	if err != nil {
		logger.Error("error on init db: ", err.Error())
		return nil, err
	}

	userService := user.NewUserService(userRepository)

	ctx, cancel := context.WithCancel(context.Background())

	accrualService := accrual.NewAccrualService(ctx, logger, cfg, userRepository)
	orderService := order.NewOrderService(logger, cfg, userRepository, accrualService)

	userHandler := user2.NewUserHandler(logger, userService)

	orderHandler := order2.NewOrderHandler(logger, orderService)

	r := router.NewRouter(userHandler, orderHandler)

	app := &App{
		Router:         r,
		logger:         logger,
		config:         cfg,
		accrualService: accrualService,
		Server: &http.Server{
			Addr:    cfg.RunAddress,
			Handler: r,
		},
	}

	app.Server.RegisterOnShutdown(cancel)

	return app, nil
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.logger.Error("Error shutting down server: ", err)
	} else {
		a.logger.Info("Server stopped gracefully")
	}
}

func (a *App) Run() {
	a.logger.Info(fmt.Sprintf("Server starting on %s", a.config.RunAddress))

	a.accrualService.Run()

	if err := a.Server.ListenAndServe(); err != nil {
		a.Stop()
	}
}
