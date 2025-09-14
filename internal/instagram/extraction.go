package instagram

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"qwiklip/internal/models"
)

// extractJSONData tries different patterns to extract JSON data from HTML
func (c *Client) extractJSONData(html string) (map[string]interface{}, error) {
	jsonPatterns := []string{
		`<script type="application/json" data-sjs>(.*?)</script>`,
		`window\.__additionalDataLoaded\('.*?',(.*?)\);`,
		`<script type="text/javascript">window\._sharedData = (.*?);</script>`,
		`window\.__APOLLO_STATE__ = (.*?);</script>`,
		`window\.__INITIAL_DATA__ = (.*?);</script>`,
		`^\{"items":`,                      // Direct JSON pattern
		`\{"graphql":\{"shortcode_media":`, // GraphQL structure pattern
	}

	c.logger.Debug("Trying JSON extraction patterns", "count", len(jsonPatterns))

	for i, pattern := range jsonPatterns {
		// Special handling for direct JSON pattern
		if pattern == `^\{"items":` {
			if strings.TrimSpace(html)[:9] == `{"items":` {
				c.logger.Debug("Direct JSON pattern matched, attempting JSON parse")
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(html), &jsonData); err == nil {
					c.logger.Info("Successfully extracted JSON data using direct JSON pattern")
					c.logJSONKeys(jsonData)
					return jsonData, nil
				} else {
					c.logger.Warn("Direct JSON parse failed", "error", err)
				}
			} else {
				c.logger.Debug("Direct JSON pattern does not match")
			}
			continue
		}

		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			c.logger.Debug("Pattern matched, attempting JSON parse", "pattern_index", i+1)
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &jsonData); err == nil {
				c.logger.Info("Successfully extracted JSON data", "pattern_index", i+1)
				c.logJSONKeys(jsonData)
				return jsonData, nil
			} else {
				c.logger.Warn("JSON parse failed", "pattern_index", i+1, "error", err)
			}
		} else {
			c.logger.Debug("Pattern did not match", "pattern_index", i+1)
		}
	}

	c.logger.Error("All JSON extraction patterns failed")
	return nil, fmt.Errorf("could not extract JSON data")
}

// logJSONKeys logs the keys found in JSON data for debugging
func (c *Client) logJSONKeys(jsonData map[string]interface{}) {
	if !c.config.Debug {
		return
	}

	keys := make([]string, 0, len(jsonData))
	for k := range jsonData {
		keys = append(keys, k)
	}
	c.logger.Debug("JSON keys found", "keys", keys)
}

// extractDirectVideoURL tries to find direct video URLs in HTML content
func (c *Client) extractDirectVideoURL(html string) (string, error) {
	// Direct video URL patterns from TypeScript implementation
	directVideoUrlPatterns := []string{
		`"video_versions":\[\{"width":\d+,"height":\d+,"url":"(https://[^"]+)"`,
		`"video_url":"(https://[^"]+)"`,
		`property="og:video" content="(https://[^"]+)"`,
		`property="og:video:secure_url" content="(https://[^"]+)"`,
		`"contentUrl":"(https://[^"]+\.mp4[^"]*)"`,
		`url":"(https://[^"]+\.mp4[^"]*)"`,
		`url":\s*"([^"]+\.mp4[^"]*)"`,
		`"url":\s*"([^"]+\.mp4[^"]*)"`,
	}

	c.logger.Debug("Trying direct video URL patterns")

	for i, pattern := range directVideoUrlPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			videoURL := c.unescapeURL(matches[1])
			c.logger.Info("Found video URL using direct pattern",
				"pattern_index", i+1,
				"url_prefix", videoURL[:min(100, len(videoURL))])
			return videoURL, nil
		}
	}

	c.logger.Error("All direct video URL patterns failed")
	return "", fmt.Errorf("no direct video URL found")
}

// unescapeURL cleans up escaped characters in URLs
func (c *Client) unescapeURL(urlStr string) string {
	// Clean up URL encoding
	replacements := map[string]string{
		`\u0025`: "%",
		`\u002F`: "/",
		`\u003A`: ":",
		`\u003F`: "?",
		`\u003D`: "=",
		`\u0026`: "&",
		`\`:      "",
	}

	result := urlStr
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// extractFallbackVideoURL tries additional fallback patterns when direct extraction fails
func (c *Client) extractFallbackVideoURL(html string, shortcode string) (string, error) {
	c.logger.Debug("Trying additional fallback video URL patterns")

	// Case-insensitive video patterns from TypeScript implementation
	fallbackVideoPatterns := []string{
		`(?i)"video_versions":\[\{[^\}]*"url":"([^"]+)"`,
		`(?i)"url":"(https://[^"]+\.mp4[^"]*)"`,
		`(?i)url":"(https://[^"]+\.mp4[^"]*)"`,
		`(?i)"url":\s*"([^"]+\.mp4[^"]*)"`,
		`(?i)url:\s*"([^"]+\.mp4[^"]*)"`,
		`(?i)"contentUrl":"([^"]+\.mp4[^"]*)"`,
	}

	for i, pattern := range fallbackVideoPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			videoURL := c.unescapeURL(matches[1])
			c.logger.Info("Found video URL with fallback pattern",
				"pattern_index", i+1,
				"url_prefix", videoURL[:min(100, len(videoURL))])
			return videoURL, nil
		}
	}

	// Try PolarisPostRootQueryRelayPreloader extraction (from TypeScript)
	c.logger.Debug("Trying PolarisPostRootQueryRelayPreloader extraction")
	preloaderPattern := `PolarisPostRootQueryRelayPreloader_[^"]+",(\{"__bbox":\{"complete":true,"result":\{"data":\{"xdt_api__v1__media__shortcode__web_info":\{"items":\[\{[^\}]+\}\]\}\}\}\}\})`
	preloaderRe := regexp.MustCompile(preloaderPattern)
	preloaderMatches := preloaderRe.FindStringSubmatch(html)

	if len(preloaderMatches) > 1 {
		var preloaderData map[string]interface{}
		if err := json.Unmarshal([]byte(preloaderMatches[1]), &preloaderData); err == nil {
			// Navigate through the nested structure
			if bbox, ok := preloaderData["__bbox"].(map[string]interface{}); ok {
				if result, ok := bbox["result"].(map[string]interface{}); ok {
					if data, ok := result["data"].(map[string]interface{}); ok {
						if apiData, ok := data["xdt_api__v1__media__shortcode__web_info"].(map[string]interface{}); ok {
							if items, ok := apiData["items"].([]interface{}); ok && len(items) > 0 {
								if media, ok := items[0].(map[string]interface{}); ok {
									if videoVersions, ok := media["video_versions"].([]interface{}); ok && len(videoVersions) > 0 {
										if version, ok := videoVersions[0].(map[string]interface{}); ok {
											if url, ok := version["url"].(string); ok {
												c.logger.Info("Found video URL in PolarisPostRootQueryRelayPreloader")
												return url, nil
											}
										}
									}
								}
							}
						}
					}
				}
			}
		} else {
			c.logger.Warn("Failed to parse PolarisPostRootQueryRelayPreloader data", "error", err)
		}
	}

	c.logger.Error("All fallback video URL patterns failed")
	return "", fmt.Errorf("no fallback video URL found")
}

// parseMediaInfo parses the JSON data to extract media information
func (c *Client) parseMediaInfo(jsonData map[string]interface{}, shortcode string) (*models.InstagramMediaInfo, error) {
	c.logger.Debug("Starting JSON parsing", "shortcode", shortcode)

	mediaInfo := &models.InstagramMediaInfo{
		FileName: fmt.Sprintf("%s.mp4", shortcode),
	}

	// Try different JSON structures to find the video URL
	c.logger.Debug("Searching for video URL in JSON data")
	videoURL := c.findVideoURL(jsonData, shortcode)
	if videoURL == "" {
		c.logger.Error("No video URL found in any JSON structure")
		return nil, fmt.Errorf("could not find video URL in Instagram response")
	}

	c.logger.Info("Found video URL in JSON data")
	mediaInfo.VideoURL = videoURL

	// Try to extract additional metadata
	c.logger.Debug("Extracting additional metadata")
	c.extractMetadata(jsonData, mediaInfo)

	if mediaInfo.Username != "" || mediaInfo.Caption != "" {
		c.logger.Debug("Metadata extracted",
			"username", mediaInfo.Username,
			"caption_length", len(mediaInfo.Caption))
	}

	return mediaInfo, nil
}
