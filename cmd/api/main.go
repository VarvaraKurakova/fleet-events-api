package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/VarvaraKurakova/fleet-events-api/internal/app"
	"github.com/VarvaraKurakova/fleet-events-api/internal/config"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()

	if err := app.RunAPI(cfg, logger); err != nil {
		logger.Error("api stopped with error", "error", err)
		os.Exit(1)
	}
}
