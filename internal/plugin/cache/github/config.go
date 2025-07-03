package github

import (
	"fmt"
	"runtime"
	"strings"
)

type GitHubConfig struct {
	// Repository in format "owner/repo" (e.g., "Schumann-IT/dehydrated-api-metadata-plugin-netscaler")
	Repository string `yaml:"repository"`

	// Version tag to use (e.g., "v1.0.0", "latest")
	// If not specified, defaults to "latest"
	Version string `yaml:"version"`

	// Platform to download (e.g., "linux-amd64", "darwin-amd64")
	// If not specified, will be auto-detected
	Platform string `yaml:"platform"`
}

func (c GitHubConfig) getPlatform() string {
	if c.Platform == "" {
		c.Platform = fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	}

	return c.Platform
}

func (c GitHubConfig) getOrg() string {
	parts := strings.Split(c.Repository, "/")
	if len(parts) != 2 {
		panic("invalid GitHub repository format, expected 'owner/repo': " + c.Repository)
	}
	return parts[0]
}

func (c GitHubConfig) getName() string {
	parts := strings.Split(c.Repository, "/")
	if len(parts) != 2 {
		panic("invalid GitHub repository format, expected 'owner/repo': " + c.Repository)
	}
	return parts[1]
}

func (c GitHubConfig) getVersion() string {
	if c.Version == "" {
		c.Version = "latest"
	}

	return c.Version
}

func (c *GitHubConfig) getReleaseUrl() string {
	if c.getVersion() == "latest" {
		return fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", c.Repository)
	}

	return fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", c.Repository, c.getVersion())
}
