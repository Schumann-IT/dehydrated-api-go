package github

import "time"

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Assets      []GitHubAsset `json:"assets"`
	PublishedAt time.Time     `json:"published_at"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}
