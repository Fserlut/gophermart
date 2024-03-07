package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Fserlut/gophermart/internal/app"
	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/logger"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.SetupLogger()

	application, err := app.CreateApp(log, cfg)

	if err != nil {
		log.Error("create app error")
		return
	}

	go func() {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
		<-stopChan

		application.Stop()
	}()

	application.Run()
}
