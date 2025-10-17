package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"qwiklip/internal/config"
	"qwiklip/internal/instagram"
	"qwiklip/internal/server"
)

// Version information set at build time
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	// Print version and exit if requested
	if *versionFlag {
		fmt.Printf("Qwiklip %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", buildTime)
		os.Exit(0)
	}

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
	versionInfo := &server.VersionInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
	srv, err := server.New(cfg, igClient, logger, versionInfo)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

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
