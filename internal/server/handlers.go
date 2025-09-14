package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qwiklip/internal/models"
)

// handleReel handles requests to /reel/{shortcode}
func (s *Server) handleReel(w http.ResponseWriter, r *http.Request) {
	instagramURL := s.parseReelURL(r.URL.Path)
	s.logger.Info("Processing Instagram URL", "url", instagramURL, "original_path", r.URL.Path)

	mediaInfo, err := s.fetchMediaInfo(instagramURL)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	s.logMediaMetadata(mediaInfo)

	// Stream the video content
	s.logger.Info("Starting video streaming")
	s.streamVideo(w, r, mediaInfo.VideoURL, mediaInfo.FileName)
}

// parseReelURL extracts and builds the Instagram URL from the request path
func (s *Server) parseReelURL(requestPath string) string {
	path := strings.TrimPrefix(requestPath, "/")
	return fmt.Sprintf("https://www.instagram.com/%s", path)
}

// fetchMediaInfo retrieves media information with timing and error handling
func (s *Server) fetchMediaInfo(instagramURL string) (*models.InstagramMediaInfo, error) {
	start := time.Now()
	mediaInfo, err := s.client.GetMediaInfo(instagramURL)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("Failed to extract media info", "error", err, "duration", duration)
		return nil, err
	}

	s.logger.Info("Successfully extracted media info",
		"duration", duration,
		"video_url_prefix", mediaInfo.VideoURL[:min(100, len(mediaInfo.VideoURL))],
		"filename", mediaInfo.FileName)

	return mediaInfo, nil
}

// logMediaMetadata logs optional media metadata
func (s *Server) logMediaMetadata(mediaInfo *models.InstagramMediaInfo) {
	if mediaInfo.Username != "" {
		s.logger.Info("Media metadata", "username", mediaInfo.Username)
	}

	if mediaInfo.Caption != "" {
		caption := mediaInfo.Caption
		if len(caption) > 100 {
			caption = caption[:100] + "..."
		}
		s.logger.Info("Media metadata", "caption", caption)
	}
}

// streamVideo streams the video content from Instagram to the client
func (s *Server) streamVideo(w http.ResponseWriter, r *http.Request, videoURL, fileName string) {
	streamer := NewVideoStreamer(s.client, s.config.Instagram.UserAgent, s.logger)
	if err := streamer.StreamVideo(w, r, videoURL, fileName); err != nil {
		s.handleError(w, r, err)
	}
}

// handleError provides structured error handling with custom error types
func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	// Check if client accepts JSON (API-style responses)
	if s.shouldReturnJSON(r) {
		s.sendErrorResponse(w, err)
		return
	}

	// Default to HTML error page for web browsers
	var appErr *models.AppError
	if errors.As(err, &appErr) {
		s.renderError(w, appErr.HTTPStatusCode(), appErr.Message,
			fmt.Sprintf("Error type: %s", string(appErr.Type)),
			s.getErrorSuggestions(string(appErr.Type)))
		return
	}

	// Generic error
	s.renderError(w, http.StatusInternalServerError, "Internal Server Error",
		"An unexpected error occurred", []string{
			"Please try again later",
			"Contact support if the problem persists",
		})
}

// sendErrorResponse sends structured JSON error responses
func (s *Server) sendErrorResponse(w http.ResponseWriter, err error) {
	var appErr *models.AppError
	if errors.As(err, &appErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatusCode())

		response := map[string]interface{}{
			"error":  appErr.Message,
			"status": http.StatusText(appErr.HTTPStatusCode()),
			"code":   appErr.HTTPStatusCode(),
			"type":   string(appErr.Type),
		}

		json.NewEncoder(w).Encode(response)
		s.logger.Error("Request failed",
			"error", appErr.Message,
			"type", string(appErr.Type),
			"status", appErr.HTTPStatusCode())
		return
	}

	// Generic error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := map[string]interface{}{
		"error":  "Internal server error",
		"status": http.StatusText(http.StatusInternalServerError),
		"code":   http.StatusInternalServerError,
	}

	json.NewEncoder(w).Encode(response)
	s.logger.Error("Unexpected error", "error", err)
}

// shouldReturnJSON determines if the client expects JSON response
func (s *Server) shouldReturnJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	// Return JSON if explicitly requested or if no specific type is requested
	return strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "*/*") ||
		accept == ""
}

// getErrorSuggestions provides contextual error suggestions based on error type
func (s *Server) getErrorSuggestions(errorType string) []string {
	switch errorType {
	case "network_error":
		return []string{
			"Check your internet connection",
			"Try again in a few minutes",
			"The service may be temporarily unavailable",
		}
	case "invalid_url":
		return []string{
			"Verify the Instagram URL is correct",
			"Ensure the URL format is /reel/{shortcode}",
			"Check that the content still exists",
		}
	case "content_not_found":
		return []string{
			"Make sure the Instagram post is public",
			"Verify the URL is correct",
			"The content may have been deleted",
		}
	case "rate_limited":
		return []string{
			"Wait a few minutes before trying again",
			"Reduce the frequency of requests",
			"Consider upgrading your plan for higher limits",
		}
	default:
		return []string{
			"Try refreshing the page",
			"Contact support if the problem persists",
		}
	}
}

// handleHealthCheck provides a simple health check endpoint
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// serveAPIInfo provides API information when templates are not available
func (s *Server) serveAPIInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	apiInfo := map[string]interface{}{
		"service": "Qwiklip",
		"version": "1.0.0",
		"status":  "running",
		"mode":    "api-only",
		"endpoints": map[string]string{
			"GET /":          "API information",
			"GET /health":    "Health check",
			"GET /reel/{id}": "Download Instagram reel",
		},
		"server": map[string]interface{}{
			"port": s.config.Server.Port,
		},
		"templates": map[string]interface{}{
			"enabled": false,
			"reason":  "Template files not found or failed to load",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(apiInfo); err != nil {
		s.logger.Error("Failed to encode API info", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleRoot provides information about the service
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if !s.templatesEnabled {
		s.logger.Info("Templates not available, serving API information in JSON format")
		s.serveAPIInfo(w, r)
		return
	}

	data := struct {
		Port string
	}{
		Port: s.config.Server.Port,
	}

	// Execute template
	if err := s.templateSet.Index.Execute(w, data); err != nil {
		s.logger.Error("Failed to execute template", "error", err)
		s.renderError(w, http.StatusInternalServerError, "Service temporarily unavailable",
			"Template rendering failed", nil)
		return
	}
}

// renderError renders an HTML error page with enhanced error handling
func (s *Server) renderError(w http.ResponseWriter, statusCode int, message string, details string, suggestions []string) {
	if !s.templatesEnabled {
		// Fallback to JSON error response when templates are not available
		s.logger.Warn("Templates not available, serving error as JSON",
			"status_code", statusCode,
			"message", message)
		s.sendErrorResponse(w, fmt.Errorf("%s: %s", message, details))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	// Ensure suggestions are provided
	if len(suggestions) == 0 {
		suggestions = s.getDefaultSuggestions(statusCode)
	}

	// Prepare error data
	errorData := struct {
		StatusCode  int
		StatusText  string
		Message     string
		Details     string
		Suggestions []string
		Timestamp   string
	}{
		StatusCode:  statusCode,
		StatusText:  http.StatusText(statusCode),
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	// Execute error template
	if err := s.templateSet.Error.Execute(w, errorData); err != nil {
		s.logger.Error("Failed to execute error template",
			"error", err,
			"status_code", statusCode,
			"message", message)
		// Final fallback to plain text
		http.Error(w, message, statusCode)
		return
	}

	// Log the error for monitoring
	s.logger.Warn("Rendered error page",
		"status_code", statusCode,
		"message", message,
		"suggestion_count", len(suggestions))
}

// getDefaultSuggestions provides default error suggestions based on HTTP status code
func (s *Server) getDefaultSuggestions(statusCode int) []string {
	switch statusCode {
	case http.StatusBadRequest:
		return []string{
			"Check the URL format",
			"Ensure all required parameters are provided",
			"Try a different Instagram URL",
		}
	case http.StatusNotFound:
		return []string{
			"Verify the Instagram URL is correct",
			"Make sure the content is still available",
			"Check if the post is private",
		}
	case http.StatusInternalServerError:
		return []string{
			"Try again in a few minutes",
			"Check your internet connection",
			"Contact support if the problem persists",
		}
	case http.StatusTooManyRequests:
		return []string{
			"Wait a few minutes before trying again",
			"Reduce the frequency of requests",
		}
	case http.StatusBadGateway:
		return []string{
			"Instagram may be temporarily unavailable",
			"Try again later",
			"Check if the content is accessible on Instagram directly",
		}
	default:
		return []string{
			"Try refreshing the page",
			"Contact support if the problem persists",
		}
	}
}

// handleNotFound handles 404 errors for unmatched routes and root requests
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	// Handle root path specially
	if r.URL.Path == "/" {
		s.handleRoot(w, r)
		return
	}

	// Use renderError for consistent error handling
	s.renderError(w, http.StatusNotFound, "Page Not Found",
		fmt.Sprintf("The requested path '%s' does not exist", r.URL.Path), []string{
			"Check the URL for typos",
			"Go back to the home page",
			"Use the correct format: /reel/{shortcode}",
		})
}

// min is a helper function for string slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
