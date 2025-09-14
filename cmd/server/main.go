package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"qwiklip/internal/config"
	"qwiklip/internal/instagram"
	"qwiklip/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Configure structured logging
	level := getLogLevel(cfg.Logging.Level)
	var handler slog.Handler
	if cfg.Logging.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Starting Qwiklip server", "port", cfg.Server.Port)

	igClient := instagram.NewClient(&cfg.Instagram, logger)

	// Initialize HTTP server
	srv := server.New(cfg, igClient, logger)

	// Setup graceful shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start server (blocks until shutdown signal)
	if err := srv.Start(ctx); err != nil {
		slog.Error("Server shutdown with error", "error", err)
		os.Exit(1)
	}
}

func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
