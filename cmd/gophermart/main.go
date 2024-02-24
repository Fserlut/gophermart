package main

import (
	"github.com/Fserlut/gophermart/internal/app"
	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/logger"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.SetupLogger()

	server := app.CreateApp(log, cfg)

	server.Run()
}
