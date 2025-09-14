package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"qwiklip/internal/config"
	"qwiklip/internal/instagram"
	"qwiklip/internal/middleware"
)

// Server represents the HTTP server with hybrid architecture
// Combines: Template support + Multiple Instagram URL patterns + Flexible middleware
type Server struct {
	config          *config.Config
	client          *instagram.Client
	logger          *slog.Logger
	httpServer      *http.Server
	template        *template.Template // Index page template
	errorTemplate   *template.Template // Error page template
	templatesLoaded bool               // Track template loading state
}

// New creates a new server instance
func New(cfg *config.Config, client *instagram.Client, logger *slog.Logger) (*Server, error) {
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

	s := &Server{
		config: cfg,
		client: client,
		logger: logger,
	}

	// Load templates
	var err error
	s.template, err = template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load index template: %w", err)
	}
	s.errorTemplate, err = template.ParseFiles("templates/error.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load error template: %w", err)
	}

	s.templatesLoaded = true
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

// MiddlewareOptions defines which middleware to apply
type MiddlewareOptions struct {
	EnableRecovery bool // Enable error recovery middleware
	EnableLogging  bool // Enable request logging middleware
	EnableCORS     bool // Enable cross-origin resource sharing middleware
}

// Quick setup helpers for common middleware configurations
func (s *Server) withMinimalMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return s.applyMiddleware(handler, MiddlewareOptions{
		EnableRecovery: false,
		EnableLogging:  false,
		EnableCORS:     false,
	})
}

func (s *Server) withStandardMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return s.applyMiddleware(handler, MiddlewareOptions{
		EnableRecovery: true,
		EnableLogging:  true,
		EnableCORS:     true,
	})
}

// applyMiddleware allows granular middleware selection for future customization
// This provides the flexibility of the previous version while maintaining clean chaining
func (s *Server) applyMiddleware(handler http.HandlerFunc, options MiddlewareOptions) http.HandlerFunc {
	result := handler

	// Apply middleware in correct order (outermost to innermost)
	if options.EnableRecovery {
		result = middleware.RecoveryMiddleware(s.logger)(result)
	}
	if options.EnableLogging {
		result = middleware.LoggingMiddleware(s.logger)(result)
	}
	if options.EnableCORS {
		result = middleware.CORSMiddleware(result)
	}

	return result
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
