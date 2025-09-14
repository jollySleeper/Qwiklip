package models

// InstagramMediaInfo represents the extracted media information from Instagram
type InstagramMediaInfo struct {
	VideoURL     string `json:"videoUrl"`
	FileName     string `json:"fileName"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Caption      string `json:"caption,omitempty"`
	Username     string `json:"username,omitempty"`
}
