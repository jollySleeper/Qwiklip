package server

import (
	"net/http"
	"qwiklip/internal/middleware"
	"qwiklip/web/static"
)

// Router handles HTTP route configuration and middleware setup
type Router struct {
	server *Server
	mux    *http.ServeMux
}

// NewRouter creates a new router instance
func NewRouter(server *Server) *Router {
	return &Router{
		server: server,
		mux:    http.NewServeMux(),
	}
}

// SetupRoutes configures all HTTP routes with appropriate middleware
func (r *Router) SetupRoutes() http.Handler {
	// Static files (favicon, images, etc.) - embedded in binary
	r.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(static.GetStaticFS())))

	// Health check endpoint - Minimal middleware for performance
	r.mux.HandleFunc("/health", r.server.withMinimalMiddleware(r.server.handleHealthCheck))

	// Instagram reel endpoint - Full middleware stack
	// Can also be written as: r.server.applyMiddleware(r.server.handleReel, ApplyMiddlewareOptions(middleware.WithRecovery(), middleware.WithLogging(), middleware.WithCORS()))
	r.mux.HandleFunc("/reel/", r.server.applyMiddleware(r.server.handleReel, middleware.DefaultConfig()))

	// Catch-all route for 404 handling
	r.mux.HandleFunc("/", r.server.withStandardMiddleware(r.server.handleNotFound))

	return r.mux
}

// Mux returns the underlying HTTP multiplexer for advanced routing needs
func (r *Router) Mux() *http.ServeMux {
	return r.mux
}
