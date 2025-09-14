package instagram

import (
	"fmt"
	"strings"

	"qwiklip/internal/models"
)

// findVideoURL tries different JSON structures to find the video URL
func (c *Client) findVideoURL(jsonData map[string]interface{}, shortcode string) string {
	c.logger.Debug("Checking different JSON structures for video URL")

	// Structure 1: PostPage format
	c.logger.Debug("Checking PostPage format")
	if require, ok := jsonData["require"].([]interface{}); ok {
		for _, item := range require {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["0"] == "PostPage" {
					if graphql, ok := itemMap["1"].(map[string]interface{})["graphql"].(map[string]interface{}); ok {
						if shortcodeMedia, ok := graphql["shortcode_media"].(map[string]interface{}); ok {
							if isVideo, ok := shortcodeMedia["is_video"].(bool); ok && isVideo {
								if videoURL, ok := shortcodeMedia["video_url"].(string); ok {
									c.logger.Info("Found video URL in PostPage graphql structure")
									return videoURL
								}
							} else if isVideo, ok := shortcodeMedia["is_video"].(bool); ok && !isVideo {
								c.logger.Debug("PostPage item is not a video")
							}
						} else {
							c.logger.Debug("PostPage graphql missing shortcode_media")
						}
					} else {
						c.logger.Debug("PostPage item missing graphql structure")
					}
				}
			}
		}
	}

	// Structure 2: SharedData format
	c.logger.Debug("Checking SharedData format")
	if entryData, ok := jsonData["entry_data"].(map[string]interface{}); ok {
		if postPage, ok := entryData["PostPage"].([]interface{}); ok && len(postPage) > 0 {
			if media := c.getShortcodeMedia(postPage[0]); media != nil {
				if videoURL := c.extractVideoURLFromMedia(media); videoURL != "" {
					c.logger.Info("Found video URL in SharedData entry_data structure")
					return videoURL
				}
			} else {
				c.logger.Debug("SharedData PostPage missing shortcode_media")
			}
		} else {
			c.logger.Debug("SharedData missing PostPage array")
		}
	} else {
		c.logger.Debug("JSON missing entry_data structure")
	}

	// Structure 3: Direct items format
	c.logger.Debug("Checking direct items format")
	if items, ok := jsonData["items"].([]interface{}); ok && len(items) > 0 {
		if media := items[0].(map[string]interface{}); media != nil {
			if videoURL := c.extractVideoURLFromMedia(media); videoURL != "" {
				c.logger.Info("Found video URL in direct items structure")
				return videoURL
			}
		} else {
			c.logger.Debug("Direct items[0] is not a valid media object")
		}
	} else {
		c.logger.Debug("JSON missing items array or items is empty")
	}

	// Structure 4: Apollo State format
	c.logger.Debug("Checking Apollo State format")
	if _, ok := jsonData["ROOT_QUERY"]; ok {
		c.logger.Debug("Found ROOT_QUERY, searching for media keys")
		for key, value := range jsonData {
			if strings.Contains(key, fmt.Sprintf("Media:%s", shortcode)) ||
				strings.Contains(key, fmt.Sprintf("ShortcodeMedia:%s", shortcode)) {
				c.logger.Debug("Found matching media key", "key", key)
				if media, ok := value.(map[string]interface{}); ok {
					if videoURL, ok := media["video_url"].(string); ok {
						c.logger.Info("Found video URL in Apollo state structure")
						return videoURL
					}
					if videoURL, ok := media["videoUrl"].(string); ok {
						c.logger.Info("Found video URL in Apollo state structure (camelCase)")
						return videoURL
					}
					c.logger.Debug("Apollo media object missing video_url fields")
				} else {
					c.logger.Debug("Apollo media value is not a valid object")
				}
			}
		}
		c.logger.Debug("No matching media keys found in Apollo state")
	} else {
		c.logger.Debug("JSON missing ROOT_QUERY (Apollo state)")
	}

	// Structure 5: Direct API response format
	c.logger.Debug("Checking direct API response format")
	if graphql, ok := jsonData["graphql"].(map[string]interface{}); ok {
		if media := c.getShortcodeMedia(graphql); media != nil {
			if videoURL := c.extractVideoURLFromMedia(media); videoURL != "" {
				c.logger.Info("Found video URL in direct API response")
				return videoURL
			}
		} else {
			c.logger.Debug("Direct API response missing shortcode_media")
		}
	} else {
		c.logger.Debug("JSON missing graphql structure")
	}

	c.logger.Error("No video URL found in any JSON structure")
	return ""
}

// Helper function to get shortcode media from various structures
func (c *Client) getShortcodeMedia(data interface{}) map[string]interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if shortcodeMedia, ok := dataMap["shortcode_media"].(map[string]interface{}); ok {
			return shortcodeMedia
		}
	}
	return nil
}

// Helper function to extract video URL from media object
func (c *Client) extractVideoURLFromMedia(media map[string]interface{}) string {
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
func (c *Client) extractMetadata(jsonData map[string]interface{}, mediaInfo *models.InstagramMediaInfo) {
	// Try to find username and caption from various structures
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
