package logger

import (
	"log/slog"
	"os"
)

// *CustomLogger
func SetupLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
}
