package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	MobileUserAgent  = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"
)

type InstagramMediaInfo struct {
	VideoURL     string `json:"videoUrl"`
	FileName     string `json:"fileName"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Caption      string `json:"caption,omitempty"`
	Username     string `json:"username,omitempty"`
}

type InstagramClient struct {
	client *http.Client
	debug  bool
}

func NewInstagramClient() *InstagramClient {
	return &InstagramClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: false,
	}
}

func (ic *InstagramClient) SetDebug(debug bool) {
	ic.debug = debug
}

// saveDebugContent saves HTML content for debugging
func (ic *InstagramClient) saveDebugContent(shortcode, content string) {
	if !ic.debug {
		return
	}

	filename := fmt.Sprintf("debug-%s-%d.html", shortcode, time.Now().Unix())
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("❌ Failed to save debug content: %v\n", err)
	} else {
		fmt.Printf("💾 Debug HTML saved to: %s\n", filename)
	}
}

// ExtractShortcode extracts the shortcode from an Instagram URL
func (ic *InstagramClient) ExtractShortcode(urlStr string) (string, error) {
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

// GetMediaInfo extracts media information from an Instagram URL
func (ic *InstagramClient) GetMediaInfo(urlStr string) (*InstagramMediaInfo, error) {
	fmt.Printf("🔍 Starting Instagram media extraction for: %s\n", urlStr)

	shortcode, err := ic.ExtractShortcode(urlStr)
	if err != nil {
		fmt.Printf("❌ Failed to extract shortcode: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ Extracted shortcode: %s\n", shortcode)

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

	fmt.Printf("🌐 Trying different URL formats and user agents...\n")

	// Try each URL format until one works
	for i, format := range urlFormats {
		userAgentType := "Desktop"
		if format.userAgent == MobileUserAgent {
			userAgentType = "Mobile"
		}

		fmt.Printf("🔄 Attempt %d: %s URL with %s user agent\n", i+1,
			format.url[:min(50, len(format.url))], userAgentType)

		req, err := http.NewRequest("GET", format.url, nil)
		if err != nil {
			fmt.Printf("❌ Failed to create request: %v\n", err)
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
		resp, err := ic.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Failed to fetch after %v: %v\n", duration, err)
			continue
		}

		fmt.Printf("📡 Response received in %v with status %d\n", duration, resp.StatusCode)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response = resp
			successURL = format.url
			fmt.Printf("✅ Successfully fetched content from: %s\n", successURL)
			break
		}

		fmt.Printf("⚠️  Status %d received, trying next format...\n", resp.StatusCode)
		resp.Body.Close()
	}

	if response == nil || successURL == "" {
		fmt.Printf("❌ All URL formats failed for shortcode: %s\n", shortcode)
		return nil, fmt.Errorf("failed to fetch Instagram content for shortcode: %s", shortcode)
	}

	defer response.Body.Close()

	fmt.Printf("📊 Content-Length: %s bytes\n", response.Header.Get("Content-Length"))
	fmt.Printf("📝 Content-Type: %s\n", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("📄 HTML content length: %d characters\n", len(body))

	// Save debug content if debug mode is enabled
	ic.saveDebugContent(shortcode, string(body))

	// Try different patterns to extract JSON data
	fmt.Printf("🔍 Attempting JSON data extraction...\n")
	jsonData, err := ic.extractJSONData(string(body))
	if err != nil {
		fmt.Printf("⚠️  JSON extraction failed, trying direct video URL extraction...\n")
		// Try to find direct video URLs in the HTML content
		videoURL, err := ic.extractDirectVideoURL(string(body))
		if err != nil {
			fmt.Printf("❌ Direct video URL extraction also failed: %v\n", err)
			fmt.Printf("🔍 Searching for video URL patterns in HTML...\n")

			// Let's search for any video-related patterns in the HTML
			videoPatterns := []string{
				`"video_url":"(https://[^"]+)"`,
				`"video_versions":\[{[^}]*"url":"(https://[^"]+)"`,
				`"contentUrl":"(https://[^"]+\.mp4[^"]*)"`,
				`property="og:video" content="(https://[^"]+)"`,
				// Additional patterns from TypeScript (case-insensitive)
				`(?i)"video_versions":\[\{[^\}]*"url":"([^"]+)"`,
				`(?i)"url":"(https://[^"]+\.mp4[^"]*)"`,
				`(?i)url":"(https://[^"]+\.mp4[^"]*)"`,
				`(?i)"url":\s*"([^"]+\.mp4[^"]*)"`,
				`(?i)url:\s*"([^"]+\.mp4[^"]*)"`,
				`(?i)"contentUrl":"([^"]+\.mp4[^"]*)"`,
			}

			fmt.Printf("🔍 Checking for video patterns in HTML...\n")
			for _, pattern := range videoPatterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindAllStringSubmatch(string(body), -1)
				if len(matches) > 0 {
					fmt.Printf("📍 Found %d matches for pattern: %s\n", len(matches), pattern)
					for i, match := range matches {
						if len(match) > 1 {
							videoURL := ic.unescapeURL(match[1])
							fmt.Printf("   Match %d: %s...\n", i+1, videoURL[:min(100, len(videoURL))])
							if strings.Contains(videoURL, ".mp4") || strings.Contains(videoURL, "instagram") {
								fmt.Printf("✅ Found video URL in HTML content: %s\n", videoURL[:min(100, len(videoURL))]+"...")
								return &InstagramMediaInfo{
									VideoURL: videoURL,
									FileName: fmt.Sprintf("%s.mp4", shortcode),
								}, nil
							}
						}
					}
				}
			}

			// Try PolarisPostRootQueryRelayPreloader extraction (from TypeScript)
			fmt.Printf("🔍 Checking for PolarisPostRootQueryRelayPreloader...\n")
			preloaderPattern := `PolarisPostRootQueryRelayPreloader_[^"]+",(\{"__bbox":\{"complete":true,"result":\{"data":\{"xdt_api__v1__media__shortcode__web_info":\{"items":\[\{[^\}]+\}\]\}\}\}\}\})`
			preloaderRe := regexp.MustCompile(preloaderPattern)
			preloaderMatches := preloaderRe.FindStringSubmatch(string(body))
			if len(preloaderMatches) > 1 {
				var preloaderData map[string]interface{}
				if err := json.Unmarshal([]byte(preloaderMatches[1]), &preloaderData); err == nil {
					// Navigate to items array
					if bbox, ok := preloaderData["__bbox"].(map[string]interface{}); ok {
						if result, ok := bbox["result"].(map[string]interface{}); ok {
							if data, ok := result["data"].(map[string]interface{}); ok {
								if api, ok := data["xdt_api__v1__media__shortcode__web_info"].(map[string]interface{}); ok {
									if items, ok := api["items"].([]interface{}); ok && len(items) > 0 {
										if item, ok := items[0].(map[string]interface{}); ok {
											if videoVersions, ok := item["video_versions"].([]interface{}); ok && len(videoVersions) > 0 {
												if version, ok := videoVersions[0].(map[string]interface{}); ok {
													if url, ok := version["url"].(string); ok {
														fmt.Printf("✅ Found video URL in PolarisPostRootQueryRelayPreloader\n")
														return &InstagramMediaInfo{
															VideoURL: url,
															FileName: fmt.Sprintf("%s.mp4", shortcode),
														}, nil
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			return nil, fmt.Errorf("could not extract video URL: %v", err)
		}

		fmt.Printf("✅ Found direct video URL: %s\n", videoURL[:min(100, len(videoURL))]+"...")

		return &InstagramMediaInfo{
			VideoURL: videoURL,
			FileName: fmt.Sprintf("%s.mp4", shortcode),
		}, nil
	}

	fmt.Printf("✅ Successfully extracted JSON data\n")

	// Parse the JSON data to find the video URL
	fmt.Printf("🔍 Parsing JSON data for video URL...\n")
	mediaInfo, err := ic.parseMediaInfo(jsonData, shortcode)
	if err != nil {
		fmt.Printf("❌ Failed to parse media info: %v\n", err)
		return nil, fmt.Errorf("failed to parse media info: %v", err)
	}

	fmt.Printf("✅ Successfully completed media extraction\n")
	return mediaInfo, nil
}

// extractJSONData tries different patterns to extract JSON data from HTML
func (ic *InstagramClient) extractJSONData(html string) (map[string]interface{}, error) {
	jsonPatterns := []string{
		`<script type="application/json" data-sjs>(.*?)</script>`,
		`window\.__additionalDataLoaded\('.*?',(.*?)\);`,
		`<script type="text/javascript">window\._sharedData = (.*?);</script>`,
		`window\.__APOLLO_STATE__ = (.*?);</script>`,
		`window\.__INITIAL_DATA__ = (.*?);</script>`,
		`^\{"items":`,                      // Direct JSON pattern - MISSING IN GO
		`\{"graphql":\{"shortcode_media":`, // GraphQL structure pattern - MISSING IN GO
	}

	fmt.Printf("🔍 Trying %d JSON extraction patterns...\n", len(jsonPatterns))

	for i, pattern := range jsonPatterns {
		// Special handling for direct JSON pattern (like TypeScript)
		if pattern == `^\{"items":` {
			if strings.TrimSpace(html)[:9] == `{"items":` {
				fmt.Printf("📋 Direct JSON pattern matched, attempting JSON parse...\n")
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(html), &jsonData); err == nil {
					fmt.Printf("✅ Successfully extracted JSON data using direct JSON pattern\n")
					fmt.Printf("📊 JSON keys found: ")
					keys := make([]string, 0, len(jsonData))
					for k := range jsonData {
						keys = append(keys, k)
					}
					fmt.Printf("%v\n", keys)
					return jsonData, nil
				} else {
					fmt.Printf("❌ Direct JSON parse failed: %v\n", err)
				}
			} else {
				fmt.Printf("❌ Direct JSON pattern does not match\n")
			}
			continue
		}

		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			fmt.Printf("📋 Pattern %d matched, attempting JSON parse...\n", i+1)
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &jsonData); err == nil {
				fmt.Printf("✅ Successfully extracted JSON data using pattern %d\n", i+1)
				fmt.Printf("📊 JSON keys found: ")
				keys := make([]string, 0, len(jsonData))
				for k := range jsonData {
					keys = append(keys, k)
				}
				fmt.Printf("%v\n", keys)
				return jsonData, nil
			} else {
				fmt.Printf("❌ JSON parse failed for pattern %d: %v\n", i+1, err)
			}
		} else {
			fmt.Printf("❌ Pattern %d did not match\n", i+1)
		}
	}

	fmt.Printf("❌ All JSON extraction patterns failed\n")
	return nil, fmt.Errorf("could not extract JSON data")
}

// extractDirectVideoURL tries to find direct video URLs in HTML content
func (ic *InstagramClient) extractDirectVideoURL(html string) (string, error) {
	// New patterns for the updated Instagram page structure
	newVideoPatterns := []string{
		// Direct MP4 URLs found in the HTML
		`"https:\/\/instagram\.fbom\d+-\d+\.fna\.fbcdn\.net\/[^"]*\.mp4[^"]*"`,
		// Base64 encoded URLs
		`https:\/\/instagram\.fbom\d+-\d+\.fna\.fbcdn\.net\/[^"]*\.mp4[^"]*`,
		// Look for video URLs in script tags
		`data:application/x-javascript;[^"]*base64,([^"]*)`,
	}

	fmt.Printf("🔍 Trying new video URL patterns for updated Instagram structure...\n")

	// First, try to find direct MP4 URLs
	for i, pattern := range newVideoPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(html, -1)
		if len(matches) > 0 {
			fmt.Printf("📍 Pattern %d found %d potential matches\n", i+1, len(matches))

			// Process matches to extract clean URLs
			for _, match := range matches {
				// Remove quotes if present
				cleanURL := strings.Trim(match, `"`)
				if strings.Contains(cleanURL, ".mp4") && strings.Contains(cleanURL, "instagram.fbom") {
					// Unescape URL encoding
					videoURL := ic.unescapeURL(cleanURL)
					fmt.Printf("✅ Found video URL: %s\n", videoURL[:min(100, len(videoURL))]+"...")
					return videoURL, nil
				}
			}
		}
	}

	// Fallback to old patterns
	oldVideoPatterns := []string{
		`"video_versions":\[\{"width":\d+,"height":\d+,"url":"(https://[^"]+)"`,
		`"video_url":"(https://[^"]+)"`,
		`property="og:video" content="(https://[^"]+)"`,
		`property="og:video:secure_url" content="(https://[^"]+)"`,
		`"contentUrl":"(https://[^"]+\.mp4[^"]*)"`,
	}

	fmt.Printf("🔄 Falling back to old video URL patterns...\n")

	for i, pattern := range oldVideoPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			videoURL := ic.unescapeURL(matches[1])
			fmt.Printf("✅ Found video URL using fallback pattern %d: %s\n",
				i+1, videoURL[:min(100, len(videoURL))]+"...")
			return videoURL, nil
		}
	}

	fmt.Printf("❌ All video URL patterns failed\n")
	return "", fmt.Errorf("no direct video URL found")
}

// unescapeURL cleans up escaped characters in URLs
func (ic *InstagramClient) unescapeURL(urlStr string) string {
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

// parseMediaInfo parses the JSON data to extract media information
func (ic *InstagramClient) parseMediaInfo(jsonData map[string]interface{}, shortcode string) (*InstagramMediaInfo, error) {
	fmt.Printf("🔍 Starting JSON parsing for shortcode: %s\n", shortcode)

	mediaInfo := &InstagramMediaInfo{
		FileName: fmt.Sprintf("%s.mp4", shortcode),
	}

	// Try different JSON structures to find the video URL
	fmt.Printf("🎯 Searching for video URL in JSON data...\n")
	videoURL := ic.findVideoURL(jsonData, shortcode)
	if videoURL == "" {
		fmt.Printf("❌ No video URL found in any JSON structure\n")
		return nil, fmt.Errorf("could not find video URL in Instagram response")
	}

	fmt.Printf("✅ Found video URL in JSON data\n")
	mediaInfo.VideoURL = videoURL

	// Try to extract additional metadata
	fmt.Printf("📋 Extracting additional metadata...\n")
	ic.extractMetadata(jsonData, mediaInfo)

	if mediaInfo.Username != "" || mediaInfo.Caption != "" {
		fmt.Printf("📊 Metadata extracted - Username: %s, Caption length: %d\n",
			mediaInfo.Username, len(mediaInfo.Caption))
	}

	return mediaInfo, nil
}

// findVideoURL tries different JSON structures to find the video URL
func (ic *InstagramClient) findVideoURL(jsonData map[string]interface{}, shortcode string) string {
	fmt.Printf("🔍 Checking different JSON structures for video URL...\n")

	// Structure 1: PostPage format
	fmt.Printf("📋 Checking PostPage format...\n")
	if require, ok := jsonData["require"].([]interface{}); ok {
		for _, item := range require {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["0"] == "PostPage" {
					if graphql, ok := itemMap["1"].(map[string]interface{})["graphql"].(map[string]interface{}); ok {
						if shortcodeMedia, ok := graphql["shortcode_media"].(map[string]interface{}); ok {
							if isVideo, ok := shortcodeMedia["is_video"].(bool); ok && isVideo {
								if videoURL, ok := shortcodeMedia["video_url"].(string); ok {
									fmt.Printf("✅ Found video URL in PostPage graphql structure\n")
									return videoURL
								}
							} else if isVideo, ok := shortcodeMedia["is_video"].(bool); ok && !isVideo {
								fmt.Printf("⚠️  PostPage item is not a video\n")
							}
						} else {
							fmt.Printf("❌ PostPage graphql missing shortcode_media\n")
						}
					} else {
						fmt.Printf("❌ PostPage item missing graphql structure\n")
					}
				}
			}
		}
	}

	// Structure 2: SharedData format
	fmt.Printf("📋 Checking SharedData format...\n")
	if entryData, ok := jsonData["entry_data"].(map[string]interface{}); ok {
		if postPage, ok := entryData["PostPage"].([]interface{}); ok && len(postPage) > 0 {
			if media := ic.getShortcodeMedia(postPage[0]); media != nil {
				if videoURL := ic.extractVideoURLFromMedia(media); videoURL != "" {
					fmt.Printf("✅ Found video URL in SharedData entry_data structure\n")
					return videoURL
				}
			} else {
				fmt.Printf("❌ SharedData PostPage missing shortcode_media\n")
			}
		} else {
			fmt.Printf("❌ SharedData missing PostPage array\n")
		}
	} else {
		fmt.Printf("❌ JSON missing entry_data structure\n")
	}

	// Structure 3: Direct items format
	fmt.Printf("📋 Checking direct items format...\n")
	if items, ok := jsonData["items"].([]interface{}); ok && len(items) > 0 {
		if media := items[0].(map[string]interface{}); media != nil {
			if videoURL := ic.extractVideoURLFromMedia(media); videoURL != "" {
				fmt.Printf("✅ Found video URL in direct items structure\n")
				return videoURL
			}
		} else {
			fmt.Printf("❌ Direct items[0] is not a valid media object\n")
		}
	} else {
		fmt.Printf("❌ JSON missing items array or items is empty\n")
	}

	// Structure 4: Apollo State format
	fmt.Printf("📋 Checking Apollo State format...\n")
	if _, ok := jsonData["ROOT_QUERY"]; ok {
		fmt.Printf("🔍 Found ROOT_QUERY, searching for media keys...\n")
		for key, value := range jsonData {
			if strings.Contains(key, fmt.Sprintf("Media:%s", shortcode)) ||
				strings.Contains(key, fmt.Sprintf("ShortcodeMedia:%s", shortcode)) {
				fmt.Printf("📍 Found matching media key: %s\n", key)
				if media, ok := value.(map[string]interface{}); ok {
					if videoURL, ok := media["video_url"].(string); ok {
						fmt.Printf("✅ Found video URL in Apollo state structure\n")
						return videoURL
					}
					if videoURL, ok := media["videoUrl"].(string); ok {
						fmt.Printf("✅ Found video URL in Apollo state structure (camelCase)\n")
						return videoURL
					}
					fmt.Printf("❌ Apollo media object missing video_url fields\n")
				} else {
					fmt.Printf("❌ Apollo media value is not a valid object\n")
				}
			}
		}
		fmt.Printf("❌ No matching media keys found in Apollo state\n")
	} else {
		fmt.Printf("❌ JSON missing ROOT_QUERY (Apollo state)\n")
	}

	// Structure 5: Direct API response format
	fmt.Printf("📋 Checking direct API response format...\n")
	if graphql, ok := jsonData["graphql"].(map[string]interface{}); ok {
		if media := ic.getShortcodeMedia(graphql); media != nil {
			if videoURL := ic.extractVideoURLFromMedia(media); videoURL != "" {
				fmt.Printf("✅ Found video URL in direct API response\n")
				return videoURL
			}
		} else {
			fmt.Printf("❌ Direct API response missing shortcode_media\n")
		}
	} else {
		fmt.Printf("❌ JSON missing graphql structure\n")
	}

	fmt.Printf("❌ No video URL found in any JSON structure\n")
	return ""
}

// Helper function to get shortcode media from various structures
func (ic *InstagramClient) getShortcodeMedia(data interface{}) map[string]interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if shortcodeMedia, ok := dataMap["shortcode_media"].(map[string]interface{}); ok {
			return shortcodeMedia
		}
	}
	return nil
}

// Helper function to extract video URL from media object
func (ic *InstagramClient) extractVideoURLFromMedia(media map[string]interface{}) string {
	if isVideo, ok := media["is_video"].(bool); ok && isVideo {
		if videoURL, ok := media["video_url"].(string); ok {
			return videoURL
		}
	}

	if videoVersions, ok := media["video_versions"].([]interface{}); ok && len(videoVersions) > 0 {
		if version, ok := videoVersions[0].(map[string]interface{}); ok {
			if url, ok := version["url"].(string); ok {
				return url
			}
		}
	}

	return ""
}

// extractMetadata tries to extract additional metadata like username, caption, etc.
func (ic *InstagramClient) extractMetadata(jsonData map[string]interface{}, mediaInfo *InstagramMediaInfo) {
	// Try to find username and caption from various structures
	// This is a simplified version - you can expand this based on the JSON structures

	// Look for owner/username in different structures
	if require, ok := jsonData["require"].([]interface{}); ok {
		for _, item := range require {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if graphql, ok := itemMap["1"].(map[string]interface{})["graphql"].(map[string]interface{}); ok {
					if shortcodeMedia, ok := graphql["shortcode_media"].(map[string]interface{}); ok {
						if owner, ok := shortcodeMedia["owner"].(map[string]interface{}); ok {
							if username, ok := owner["username"].(string); ok {
								mediaInfo.Username = username
							}
						}
						if caption, ok := shortcodeMedia["edge_media_to_caption"].(map[string]interface{}); ok {
							if edges, ok := caption["edges"].([]interface{}); ok && len(edges) > 0 {
								if edge, ok := edges[0].(map[string]interface{}); ok {
									if node, ok := edge["node"].(map[string]interface{}); ok {
										if text, ok := node["text"].(string); ok {
											mediaInfo.Caption = text
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// min is a helper function for string slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
