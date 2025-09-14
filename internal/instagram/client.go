package instagram

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qwiklip/internal/config"
	"qwiklip/internal/models"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	MobileUserAgent  = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"
)

// Client handles Instagram media extraction
type Client struct {
	httpClient *http.Client
	config     *config.InstagramConfig
	logger     *slog.Logger
}

// NewClient creates a new Instagram client
func NewClient(cfg *config.InstagramConfig, logger *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		logger: logger,
	}
}

// GetHTTPClient returns the underlying HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// GetMediaInfo extracts media information from an Instagram URL
func (c *Client) GetMediaInfo(instagramURL string) (*models.InstagramMediaInfo, error) {
	c.logger.Info("Starting Instagram media extraction", "url", instagramURL)

	shortcode, err := c.ExtractShortcode(instagramURL)
	if err != nil {
		c.logger.Error("Failed to extract shortcode", "error", err, "url", instagramURL)
		return nil, err // Return the error directly
	}

	c.logger.Info("Extracted shortcode", "shortcode", shortcode)

	// Try different URL formats to increase success chances
	urlFormats := []struct {
		url       string
		userAgent string
	}{
		{fmt.Sprintf("https://www.instagram.com/p/%s/", shortcode), DefaultUserAgent},
		{fmt.Sprintf("https://www.instagram.com/reel/%s/", shortcode), DefaultUserAgent},
		{fmt.Sprintf("https://www.instagram.com/p/%s/", shortcode), MobileUserAgent},
		{fmt.Sprintf("https://www.instagram.com/reel/%s/?__a=1&__d=dis", shortcode), MobileUserAgent},
	}

	var response *http.Response
	var successURL string

	c.logger.Info("Trying different URL formats and user agents")

	// Try each URL format until one works
	for i, format := range urlFormats {
		userAgentType := "Desktop"
		if format.userAgent == MobileUserAgent {
			userAgentType = "Mobile"
		}

		c.logger.Debug("Attempt",
			"attempt", i+1,
			"url_format", format.url[:min(50, len(format.url))],
			"user_agent", userAgentType)

		req, err := http.NewRequest("GET", format.url, nil)
		if err != nil {
			c.logger.Error("Failed to create request", "error", err)
			continue
		}

		req.Header.Set("User-Agent", format.userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Referer", "https://www.instagram.com/")
		req.Header.Set("sec-fetch-dest", "document")
		req.Header.Set("sec-fetch-mode", "navigate")
		req.Header.Set("sec-fetch-site", "same-origin")
		req.Header.Set("sec-fetch-user", "?1")
		req.Header.Set("upgrade-insecure-requests", "1")

		start := time.Now()
		resp, err := c.httpClient.Do(req)
		duration := time.Since(start)

		if err != nil {
			c.logger.Error("Failed to fetch", "error", err, "duration", duration)
			continue
		}

		c.logger.Debug("Response received", "status", resp.StatusCode, "duration", duration)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response = resp
			successURL = format.url
			c.logger.Info("Successfully fetched content", "url", successURL)
			// Read and check the response body here
			body, err := io.ReadAll(response.Body)
			if err != nil {
				c.logger.Error("Failed to read response body", "error", err)
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			response.Body.Close() // Close the original body

			c.logger.Debug("HTML content length", "length", len(body))

			// Check if this is an Instagram 404 page
			if c.isInstagram404Page(string(body)) {
				c.logger.Warn("Detected Instagram 404 page", "shortcode", shortcode)
				return nil, models.NewNotFoundError(fmt.Sprintf("Instagram content with shortcode '%s'", shortcode))
			}

			// Save debug content if debug mode is enabled
			c.saveDebugContent(shortcode, string(body))

			// Create a new ReadCloser for the response body since we consumed it
			response.Body = io.NopCloser(strings.NewReader(string(body)))
			break
		}

		// Handle error status codes that indicate we should stop trying
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			if resp.StatusCode == 404 {
				c.logger.Warn("Content not found (404), stopping attempts", "url", format.url)
				return nil, models.NewNotFoundError("Instagram content")
			} else if resp.StatusCode == 429 {
				c.logger.Warn("Rate limited (429), stopping attempts", "url", format.url)
				return nil, models.NewRateLimitedError("")
			} else if resp.StatusCode >= 500 {
				c.logger.Warn("Instagram server error, stopping attempts", "status", resp.StatusCode, "url", format.url)
				return nil, models.NewNetworkError("Instagram server error", fmt.Errorf("HTTP %d", resp.StatusCode))
			} else {
				c.logger.Warn("Client error, stopping attempts", "status", resp.StatusCode, "url", format.url)
				return nil, models.NewNetworkError("Instagram client error", fmt.Errorf("HTTP %d", resp.StatusCode))
			}
		}
		resp.Body.Close()
	}

	if response == nil || successURL == "" {
		c.logger.Error("All URL formats failed", "shortcode", shortcode)
		return nil, models.NewNotFoundError(fmt.Sprintf("Instagram content with shortcode '%s'", shortcode))
	}

	defer response.Body.Close()

	c.logger.Debug("Content info",
		"content_length", response.Header.Get("Content-Length"),
		"content_type", response.Header.Get("Content-Type"))

	// Get the body content (already read and checked above)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try different patterns to extract JSON data
	c.logger.Debug("Attempting JSON data extraction")
	jsonData, err := c.extractJSONData(string(body), shortcode)
	if err != nil {
		c.logger.Warn("JSON extraction failed, trying direct video URL extraction", "error", err)
		// Try to find direct video URLs in the HTML content
		videoURL, err := c.extractDirectVideoURL(string(body))
		if err != nil {
			c.logger.Error("Direct video URL extraction also failed", "error", err)
			// Try additional fallback patterns from TypeScript implementation
			videoURL, err = c.extractFallbackVideoURL(string(body), shortcode)
			if err != nil {
				c.logger.Error("Fallback video URL extraction also failed", "error", err)
				return nil, err // Return the error directly
			}
		}

		c.logger.Info("Found direct video URL", "url_prefix", videoURL[:min(100, len(videoURL))])

		return &models.InstagramMediaInfo{
			VideoURL: videoURL,
			FileName: fmt.Sprintf("%s.mp4", shortcode),
		}, nil
	}

	c.logger.Info("Successfully extracted JSON data")

	// Parse the JSON data to find the video URL
	c.logger.Debug("Parsing JSON data for video URL")
	mediaInfo, err := c.parseMediaInfo(jsonData, shortcode)
	if err != nil {
		c.logger.Error("Failed to parse media info", "error", err, "error_type", fmt.Sprintf("%T", err))
		return nil, err // Return the error directly without wrapping
	}

	c.logger.Info("Successfully completed media extraction")
	return mediaInfo, nil
}

// ExtractShortcode extracts the shortcode from an Instagram URL
func (c *Client) ExtractShortcode(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "instagram.com") {
		return "", fmt.Errorf("not an Instagram URL: %s", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", urlStr)
	}

	path := strings.TrimSuffix(parsedURL.Path, "/")
	segments := strings.Split(path, "/")

	if len(segments) >= 3 {
		pathType := segments[len(segments)-2]
		if pathType == "p" || pathType == "reel" || pathType == "tv" {
			return segments[len(segments)-1], nil
		}
	}

	return "", fmt.Errorf("could not extract shortcode from URL: %s", urlStr)
}

// isInstagram404Page checks if the HTML content indicates an Instagram 404 page
func (c *Client) isInstagram404Page(html string) bool {
	c.logger.Info("Checking for Instagram 404 page indicators", "content_length", len(html))

	// Common indicators of Instagram 404 pages
	indicators := []string{
		"Sorry, this page isn't available",
		"The link you followed may be broken",
		"this post is unavailable",
		"content isn't available",
		"this content isn't available right now",
		"this account has been suspended",
		"this account may have been deactivated",
		"Page Not Found",
		"post not found",
		"video not found",
	}

	htmlLower := strings.ToLower(html)
	for _, indicator := range indicators {
		if strings.Contains(htmlLower, strings.ToLower(indicator)) {
			c.logger.Info("Found 404 indicator in HTML", "indicator", indicator)
			return true
		}
	}

	// Log a sample of the HTML content (only in debug mode)
	if c.config.Debug && len(html) > 200 {
		c.logger.Debug("HTML content sample", "sample", html[:200]+"...")
	}

	return false
}

// saveDebugContent saves HTML content for debugging
func (c *Client) saveDebugContent(shortcode, content string) {
	if !c.config.Debug {
		return
	}

	// Create debug directory if it doesn't exist
	debugDir := "debug"
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		c.logger.Error("Failed to create debug directory", "error", err)
		return
	}

	filename := filepath.Join(debugDir, fmt.Sprintf("debug-%s-%d.html", shortcode, time.Now().Unix()))
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		c.logger.Error("Failed to save debug content", "error", err, "filename", filename)
	}
}
