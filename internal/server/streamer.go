package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"qwiklip/internal/instagram"
)

// VideoStreamer handles video streaming from Instagram to clients
type VideoStreamer struct {
	userAgent string
	logger    *slog.Logger
	client    *instagram.Client
}

// NewVideoStreamer creates a new video streamer
func NewVideoStreamer(client *instagram.Client, userAgent string, logger *slog.Logger) *VideoStreamer {
	return &VideoStreamer{
		userAgent: userAgent,
		logger:    logger,
		client:    client,
	}
}

// StreamVideo streams video content from Instagram to the client
func (vs *VideoStreamer) StreamVideo(w http.ResponseWriter, r *http.Request, videoURL, fileName string) error {
	vs.logger.Debug("Creating request to Instagram video URL")

	req, err := vs.createVideoRequest(r.Context(), videoURL, r)
	if err != nil {
		vs.logger.Error("Failed to create video request", "error", err)
		return err
	}

	resp, err := vs.makeVideoRequest(req)
	if err != nil {
		vs.logger.Error("Failed to fetch video", "error", err)
		return err
	}
	defer resp.Body.Close()

	if err := vs.validateResponse(resp); err != nil {
		return err
	}

	vs.setResponseHeaders(w, resp)

	return vs.streamContent(w, resp.Body, fileName)
}

// createVideoRequest creates an HTTP request to fetch the video
func (vs *VideoStreamer) createVideoRequest(ctx context.Context, videoURL string, originalReq *http.Request) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		return nil, err
	}

	vs.setBrowserHeaders(req)

	// Add Range header if present in the original request (for partial content)
	if rangeHeader := originalReq.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
		vs.logger.Debug("Range request", "range", rangeHeader)
	}

	return req, nil
}

// setBrowserHeaders sets headers to mimic a browser request
func (vs *VideoStreamer) setBrowserHeaders(req *http.Request) {
	headers := map[string]string{
		"User-Agent":      vs.userAgent,
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9",
		"Referer":         "https://www.instagram.com/",
		"Origin":          "https://www.instagram.com",
		"Connection":      "keep-alive",
		"Sec-Fetch-Dest":  "video",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "cross-site",
		"Pragma":          "no-cache",
		"Cache-Control":   "no-cache",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// makeVideoRequest executes the HTTP request to Instagram
func (vs *VideoStreamer) makeVideoRequest(req *http.Request) (*http.Response, error) {
	vs.logger.Debug("Making request to Instagram CDN")
	start := time.Now()

	resp, err := vs.client.GetHTTPClient().Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}

	vs.logger.Info("Instagram CDN responded", "status", resp.StatusCode, "duration", duration)
	return resp, nil
}

// validateResponse checks if the Instagram response is valid
func (vs *VideoStreamer) validateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		vs.logger.Error("Instagram CDN returned error status", "status", resp.StatusCode)
		return fmt.Errorf("instagram server responded with status: %d", resp.StatusCode)
	}
	return nil
}

// setResponseHeaders sets appropriate headers on the client response
func (vs *VideoStreamer) setResponseHeaders(w http.ResponseWriter, resp *http.Response) {
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	// Set Content-Length if available
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
		vs.logger.Debug("Content info", "content_length", contentLength)
	}

	// Set Content-Range if available (for partial content)
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
		vs.logger.Debug("Content info", "content_range", contentRange)
	}

	// Set status code
	if resp.StatusCode == http.StatusPartialContent {
		w.WriteHeader(http.StatusPartialContent)
		vs.logger.Debug("Sending partial content response")
	} else {
		w.WriteHeader(http.StatusOK)
		vs.logger.Debug("Sending OK response")
	}
}

// streamContent streams the video content to the client with progress logging
func (vs *VideoStreamer) streamContent(w http.ResponseWriter, body io.ReadCloser, fileName string) error {
	vs.logger.Info("Starting video streaming to client")

	buffer := make([]byte, 64*1024) // 64KB buffer
	totalBytes := 0
	streamStart := time.Now()

	for {
		n, err := body.Read(buffer)
		if n > 0 {
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				vs.logger.Warn("Client disconnected during streaming", "error", writeErr)
				return nil // Client disconnect is not an error
			}
			totalBytes += n

			// Log progress for large files (every 1MB)
			if totalBytes%(1024*1024) == 0 {
				elapsed := time.Since(streamStart)
				rate := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024 // MB/s
				vs.logger.Info("Stream progress",
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
				vs.logger.Info("Successfully streamed video",
					"filename", fileName,
					"total_bytes", totalBytes,
					"rate_mbs", fmt.Sprintf("%.2f", avgRate),
					"duration", totalTime)
				return nil
			} else {
				vs.logger.Error("Error streaming video", "filename", fileName, "error", err)
				return err
			}
		}
	}
}
