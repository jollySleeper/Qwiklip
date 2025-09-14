package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"qwiklip/internal/models"
)

// handleReel handles requests to /reel/{shortcode}
func (s *Server) handleReel(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/")
	instagramURL := fmt.Sprintf("https://www.instagram.com/%s", requestPath)

	s.logger.Info("Processing Instagram URL", "url", instagramURL, "original_path", r.URL.Path)

	// Get media information from Instagram
	start := time.Now()
	mediaInfo, err := s.client.GetMediaInfo(instagramURL)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("Failed to extract media info", "error", err, "duration", duration)
		s.handleError(w, r, err)
		return
	}

	s.logger.Info("Successfully extracted media info",
		"duration", duration,
		"video_url_prefix", mediaInfo.VideoURL[:min(100, len(mediaInfo.VideoURL))],
		"filename", mediaInfo.FileName)

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

	// Stream the video content
	s.logger.Info("Starting video streaming")
	s.streamVideo(w, r, mediaInfo.VideoURL, mediaInfo.FileName)
}

// streamVideo streams the video content from Instagram to the client
func (s *Server) streamVideo(w http.ResponseWriter, r *http.Request, videoURL, fileName string) {
	s.logger.Debug("Creating request to Instagram video URL")

	// Create a new request to fetch the video
	req, err := http.NewRequestWithContext(r.Context(), "GET", videoURL, nil)
	if err != nil {
		s.logger.Error("Failed to create video request", "error", err)
		s.handleError(w, r, err)
		return
	}

	// Set headers to mimic a browser request
	req.Header.Set("User-Agent", s.config.Instagram.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.instagram.com/")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "video")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")

	// Add Range header if present in the original request (for partial content)
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
		s.logger.Debug("Range request", "range", rangeHeader)
	}

	s.logger.Debug("Making request to Instagram CDN")
	start := time.Now()

	// Make the request
	resp, err := s.client.GetHTTPClient().Do(req)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("Failed to fetch video", "error", err, "duration", duration)
		s.handleError(w, r, err)
		return
	}
	defer resp.Body.Close()

	s.logger.Info("Instagram CDN responded", "status", resp.StatusCode, "duration", duration)

	// Check if the request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.logger.Error("Instagram CDN returned error status", "status", resp.StatusCode)
		s.renderError(w, http.StatusBadGateway, "Content temporarily unavailable",
			fmt.Sprintf("Instagram server responded with status: %d", resp.StatusCode), nil)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	// Set Content-Length if available
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
		s.logger.Debug("Content info", "content_length", contentLength)
	}

	// Set Content-Range if available (for partial content)
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
		s.logger.Debug("Content info", "content_range", contentRange)
	}

	// Set status code
	if resp.StatusCode == http.StatusPartialContent {
		w.WriteHeader(http.StatusPartialContent)
		s.logger.Debug("Sending partial content response")
	} else {
		w.WriteHeader(http.StatusOK)
		s.logger.Debug("Sending OK response")
	}

	// Stream the video content to the client
	s.logger.Info("Starting video streaming to client")
	buffer := make([]byte, 64*1024) // 64KB buffer
	totalBytes := 0
	streamStart := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				s.logger.Warn("Client disconnected during streaming", "error", writeErr)
				return
			}
			totalBytes += n

			// Log progress for large files (every 1MB)
			if totalBytes%(1024*1024) == 0 {
				elapsed := time.Since(streamStart)
				rate := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024 // MB/s
				s.logger.Info("Stream progress",
					"streamed_mb", totalBytes/(1024*1024),
					"filename", fileName,
					"rate_mbs", fmt.Sprintf("%.2f", rate))
			}
		}

		if err != nil {
			if err == io.EOF {
				totalTime := time.Since(streamStart)
				avgRate := float64(0)
				if totalTime.Seconds() > 0 {
					avgRate = float64(totalBytes) / totalTime.Seconds() / 1024 / 1024 // MB/s
				}
				s.logger.Info("Successfully streamed video",
					"filename", fileName,
					"total_bytes", totalBytes,
					"rate_mbs", fmt.Sprintf("%.2f", avgRate),
					"duration", totalTime)
			} else {
				s.logger.Error("Error streaming video", "filename", fileName, "error", err)
			}
			break
		}
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

// handleRoot provides information about the service
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if s.template == nil {
		s.logger.Error("HTML template not loaded")
		s.renderError(w, http.StatusInternalServerError, "Service temporarily unavailable",
			"HTML template not available", nil)
		return
	}

	data := struct {
		Port string
	}{
		Port: s.config.Server.Port,
	}

	// Execute template
	if err := s.template.Execute(w, data); err != nil {
		s.logger.Error("Failed to execute template", "error", err)
		s.renderError(w, http.StatusInternalServerError, "Service temporarily unavailable",
			"Template rendering failed", nil)
		return
	}
}

// renderError renders an HTML error page with enhanced error handling
func (s *Server) renderError(w http.ResponseWriter, statusCode int, message string, details string, suggestions []string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	if s.errorTemplate == nil {
		// Fallback to plain text if error template is not available
		s.logger.Error("Error template not loaded, falling back to plain text",
			"status_code", statusCode,
			"message", message)
		http.Error(w, message, statusCode)
		return
	}

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
	if err := s.errorTemplate.Execute(w, errorData); err != nil {
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
