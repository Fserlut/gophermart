package main

import (
	"github.com/Fserlut/gophermart/internal/app"
	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/logger"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.SetupLogger()

	server, err := app.CreateApp(log, cfg)

	if err != nil {
		log.Error("create app error")
		return
	}

	server.Run()
}
