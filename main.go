package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	instagramClient *InstagramClient
	port            string
	debug           bool
}

func NewServer(port string, debug bool) *Server {
	client := NewInstagramClient()
	client.SetDebug(debug)
	return &Server{
		instagramClient: client,
		port:            port,
		debug:           debug,
	}
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		fmt.Printf("[%s] %s %s %s - Request started\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			clientIP)

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapper, r)

		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %d %v - Request completed\n",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			duration)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleReel handles requests to /reel/{shortcode} and /p/{shortcode}
func (s *Server) handleReel(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/")
	instagramURL := fmt.Sprintf("https://www.instagram.com/%s", requestPath)

	fmt.Printf("🔍 Processing Instagram URL: %s\n", instagramURL)
	fmt.Printf("📝 Original request path: %s\n", r.URL.Path)

	// Get media information from Instagram
	start := time.Now()
	mediaInfo, err := s.instagramClient.GetMediaInfo(instagramURL)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Failed to extract media info after %v: %v\n", duration, err)
		http.Error(w, fmt.Sprintf("Failed to get media info: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("✅ Successfully extracted media info in %v\n", duration)
	fmt.Printf("🎬 Video URL: %s\n", mediaInfo.VideoURL[:min(100, len(mediaInfo.VideoURL))]+"...")
	fmt.Printf("📁 File name: %s\n", mediaInfo.FileName)

	if mediaInfo.Username != "" {
		fmt.Printf("👤 Username: %s\n", mediaInfo.Username)
	}

	if mediaInfo.Caption != "" {
		caption := mediaInfo.Caption
		if len(caption) > 100 {
			caption = caption[:100] + "..."
		}
		fmt.Printf("📝 Caption: %s\n", caption)
	}

	// Stream the video content
	fmt.Printf("🎥 Starting video streaming...\n")
	s.streamVideo(w, r, mediaInfo.VideoURL, mediaInfo.FileName)
}

// streamVideo streams the video content from Instagram to the client
func (s *Server) streamVideo(w http.ResponseWriter, r *http.Request, videoURL, fileName string) {
	fmt.Printf("🌐 Creating request to Instagram video URL\n")

	// Create a new request to fetch the video
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create video request: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers to mimic a browser request
	req.Header.Set("User-Agent", DefaultUserAgent)
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
		fmt.Printf("📊 Range request: %s\n", rangeHeader)
	}

	fmt.Printf("📡 Making request to Instagram CDN...\n")
	start := time.Now()

	// Make the request
	resp, err := s.instagramClient.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Failed to fetch video after %v: %v\n", duration, err)
		http.Error(w, fmt.Sprintf("Failed to fetch video: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Instagram CDN responded in %v with status %d\n", duration, resp.StatusCode)

	// Check if the request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("❌ Instagram CDN returned error status: %d\n", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Instagram server responded with status: %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	// Set Content-Length if available
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
		fmt.Printf("📏 Content-Length: %s bytes\n", contentLength)
	}

	// Set Content-Range if available (for partial content)
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
		fmt.Printf("📊 Content-Range: %s\n", contentRange)
	}

	// Set status code
	if resp.StatusCode == http.StatusPartialContent {
		w.WriteHeader(http.StatusPartialContent)
		fmt.Printf("📊 Sending partial content response\n")
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Printf("📤 Sending OK response\n")
	}

	// Stream the video content to the client
	fmt.Printf("🚀 Starting video streaming to client...\n")
	buffer := make([]byte, 64*1024) // 64KB buffer
	totalBytes := 0
	streamStart := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				fmt.Printf("⚠️  Client disconnected during streaming: %v\n", writeErr)
				return
			}
			totalBytes += n

			// Log progress for large files (every 1MB)
			if totalBytes%(1024*1024) == 0 {
				elapsed := time.Since(streamStart)
				rate := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024 // MB/s
				fmt.Printf("📊 Streamed %d MB for %s (%.2f MB/s)\n",
					totalBytes/(1024*1024), fileName, rate)
			}
		}

		if err != nil {
			if err == io.EOF {
				totalTime := time.Since(streamStart)
				avgRate := float64(totalBytes) / totalTime.Seconds() / 1024 / 1024 // MB/s
				fmt.Printf("✅ Successfully streamed video: %s (%d bytes, %.2f MB/s, %v)\n",
					fileName, totalBytes, avgRate, totalTime)
			} else {
				fmt.Printf("❌ Error streaming video %s: %v\n", fileName, err)
			}
			break
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
	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Instagram Proxy Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .example { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0; }
        code { background: #e0e0e0; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Instagram Proxy Server</h1>
        <p>This server mirrors Instagram video URLs. Use the same path structure as Instagram:</p>

        <h2>Examples:</h2>

        <div class="example">
            <strong>Instagram URL:</strong><br>
            <code>https://www.instagram.com/reel/ABC123XYZ/</code><br><br>
            <strong>Proxy URL:</strong><br>
            <code>http://localhost:8080/reel/ABC123XYZ/</code>
        </div>

        <div class="example">
            <strong>Instagram URL:</strong><br>
            <code>https://www.instagram.com/p/DEF456UVW/</code><br><br>
            <strong>Proxy URL:</strong><br>
            <code>http://localhost:8080/p/DEF456UVW/</code>
        </div>

        <h2>Features:</h2>
        <ul>
            <li>✅ Direct video streaming</li>
            <li>✅ Supports reels and posts</li>
            <li>✅ Multiple extraction strategies</li>
            <li>✅ Range request support</li>
            <li>✅ Health check endpoint</li>
        </ul>

        <h2>Health Check:</h2>
        <p><a href="/health">/health</a></p>
    </div>
</body>
</html>`
	fmt.Fprint(w, html)
}

// SetupRoutes configures all the HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint with logging
	mux.HandleFunc("/health", s.loggingMiddleware(s.handleHealthCheck))

	// Root endpoint with information and logging
	mux.HandleFunc("/", s.loggingMiddleware(s.handleRoot))

	// Handle Instagram URL patterns with logging
	mux.HandleFunc("/reel/", s.loggingMiddleware(s.handleReel))
	mux.HandleFunc("/p/", s.loggingMiddleware(s.handleReel))
	mux.HandleFunc("/tv/", s.loggingMiddleware(s.handleReel))

	return mux
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := s.SetupRoutes()

	fmt.Printf("🚀 Instagram Proxy Server starting on port %s\n", s.port)
	fmt.Printf("📺 Access videos at: http://localhost:%s/reel/{shortcode}/\n", s.port)
	fmt.Printf("ℹ️  Server info at: http://localhost:%s/\n", s.port)

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Longer timeout for video streaming
		IdleTimeout:  120 * time.Second,
	}

	return server.ListenAndServe()
}

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get debug mode from environment variable
	debug := os.Getenv("DEBUG") == "true"

	// Validate port number
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid port number: %s", port)
	}

	server := NewServer(port, debug)

	// Handle graceful shutdown (optional)
	go func() {
		fmt.Println("Server started. Press Ctrl+C to stop.")
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
