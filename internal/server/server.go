package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"qwiklip/internal/config"
	"qwiklip/internal/instagram"
	"qwiklip/internal/middleware"
	"qwiklip/web/templates"
)

// VersionInfo holds version information for the application
type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

// HTTPServer defines the interface for HTTP server lifecycle management
type HTTPServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Server represents the HTTP server with hybrid architecture
// Combines: Template support + Multiple Instagram URL patterns + Flexible middleware
type Server struct {
	config           *config.Config
	client           *instagram.Client
	logger           *slog.Logger
	httpServer       *http.Server
	templateSet      *templates.TemplateSet // Parsed HTML templates (optional)
	templatesEnabled bool                   // Whether templates are available for use
	versionInfo      *VersionInfo           // Version information for templates
}

// New creates a new server instance
func New(cfg *config.Config, client *instagram.Client, logger *slog.Logger, versionInfo *VersionInfo) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if cfg.Server.Port == "" {
		return nil, errors.New("server port is required")
	}
	if client == nil {
		return nil, errors.New("instagram client cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if versionInfo == nil {
		return nil, errors.New("version info cannot be nil")
	}

	s := &Server{
		config:      cfg,
		client:      client,
		logger:      logger,
		versionInfo: versionInfo,
	}

	// Load templates (optional - server can run in API-only mode)
	templateSet, err := templates.Load()
	if err != nil {
		s.logger.Warn("Templates not available, server will run in API-only mode", "error", err)
		s.templatesEnabled = false
	} else {
		s.templateSet = templateSet
		s.templatesEnabled = true
	}

	return s, nil
}

// Start starts the HTTP server and blocks until shutdown
func (s *Server) Start(ctx context.Context) error {
	// Setup routes with middleware
	router := NewRouter(s)
	handler := router.SetupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      handler,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	// Start server in background
	go func() {
		s.logger.Info("Server starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed to start", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	s.logger.Info("Shutting down server...")

	// Graceful shutdown
	return s.gracefulShutdown()
}

// Import middleware types for cleaner usage
type MiddlewareConfig = middleware.MiddlewareConfig
type MiddlewareOption = middleware.MiddlewareOption

// Quick setup helpers for common middleware configurations
func (s *Server) withMinimalMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return s.applyMiddleware(handler, middleware.MinimalConfig())
}

func (s *Server) withStandardMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return s.applyMiddleware(handler, middleware.DefaultConfig())
}

// applyMiddleware applies middleware configuration to a handler
func (s *Server) applyMiddleware(handler http.HandlerFunc, config *MiddlewareConfig) http.HandlerFunc {
	result := handler

	// Apply middleware in correct order (outermost to innermost)
	if config.EnableRecovery {
		result = middleware.RecoveryMiddleware(s.logger)(result)
	}
	if config.EnableLogging {
		result = middleware.LoggingMiddleware(s.logger)(result)
	}
	if config.EnableCORS {
		result = middleware.CORSMiddleware(result)
	}

	return result
}

// ApplyMiddlewareOptions applies functional options to create middleware configuration
func ApplyMiddlewareOptions(opts ...MiddlewareOption) *MiddlewareConfig {
	return middleware.ApplyOptions(opts...)
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// gracefulShutdown performs graceful server shutdown
func (s *Server) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	s.logger.Info("Server exited gracefully")
	return nil
}
