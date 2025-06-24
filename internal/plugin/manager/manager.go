package manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"go.uber.org/zap"
)

// Manager handles downloading and managing plugins from GitHub
type Manager struct {
	logger         *zap.Logger
	cacheDir       string
	client         *http.Client
	downloadClient *http.Client
}

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

// NewManager creates a new plugin manager
func NewManager(logger *zap.Logger, cacheDir string) *Manager {
	if cacheDir == "" {
		// Use current working directory for cache
		wd, err := os.Getwd()
		if err != nil {
			// Fallback to temp directory if we can't get working directory
			cacheDir = filepath.Join(os.TempDir(), "dehydrated-api-plugins")
		} else {
			cacheDir = filepath.Join(wd, ".dehydrated-api-go", "plugins")
		}
	}

	return &Manager{
		logger:   logger,
		cacheDir: cacheDir,
		client: &http.Client{
			Timeout: 30 * time.Second, // Shorter timeout for API calls
		},
		downloadClient: &http.Client{
			Timeout: 10 * time.Minute, // Longer timeout for file downloads
		},
	}
}

// GetPluginPath returns the path to a plugin executable
// If it's a GitHub plugin, it will download and cache it
func (m *Manager) GetPluginPath(pluginConfig config.PluginConfig) (string, error) {
	// If Path is set, use it directly (takes precedence)
	if pluginPath := pluginConfig.GetPluginPath(); pluginPath != "" {
		return pluginPath, nil
	}

	// If GitHub is configured, download and cache the plugin
	if githubInfo := pluginConfig.GetGitHubInfo(); githubInfo != nil {
		return m.downloadGitHubPlugin(githubInfo)
	}

	return "", fmt.Errorf("no plugin path or GitHub configuration provided")
}

// DownloadGitHubPlugin downloads a plugin from GitHub using config map
func (m *Manager) DownloadGitHubPlugin(configMap map[string]any) (string, error) {
	repository, ok := configMap["repository"].(string)
	if !ok {
		return "", fmt.Errorf("repository is required for GitHub registry")
	}

	version, _ := configMap["version"].(string)
	if version == "" {
		version = "latest"
	}

	platform, _ := configMap["platform"].(string)

	// Convert to the old config format for compatibility
	oldConfig := &config.GitHubConfig{
		Repository: repository,
		Version:    version,
		Platform:   platform,
	}
	return m.downloadGitHubPlugin(oldConfig)
}

// downloadGitHubPlugin downloads a plugin from GitHub and returns its path
func (m *Manager) downloadGitHubPlugin(githubInfo *config.GitHubConfig) (string, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Determine version to download
	version := githubInfo.Version
	if version == "" {
		version = "latest"
	}

	// Determine platform
	platform := githubInfo.Platform
	if platform == "" {
		platform = m.detectPlatform()
	}

	// Parse repository to get organization and plugin name
	org, pluginName, err := m.parseRepository(githubInfo.Repository)
	if err != nil {
		return "", err
	}

	// Generate cache path using new directory structure
	cachePath := m.generateCachePath("github", org, pluginName, version, platform)

	// Check if plugin is already cached
	if m.isPluginCached(cachePath) {
		m.logger.Debug("Using cached plugin",
			zap.String("repository", githubInfo.Repository),
			zap.String("version", version),
			zap.String("platform", platform),
			zap.String("cachePath", cachePath))
		return cachePath, nil
	}

	// Download the plugin
	m.logger.Info("Downloading plugin from GitHub",
		zap.String("repository", githubInfo.Repository),
		zap.String("version", version),
		zap.String("platform", platform))

	pluginPath, err := m.downloadPlugin(githubInfo.Repository, version, platform, cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}

	// Make the plugin executable
	if err := os.Chmod(pluginPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	m.logger.Info("Plugin downloaded successfully", zap.String("path", pluginPath))
	return pluginPath, nil
}

// parseRepository parses a GitHub repository string into organization and plugin name
func (m *Manager) parseRepository(repository string) (org, pluginName string, err error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", repository)
	}
	return parts[0], parts[1], nil
}

// generateCachePath generates the cache path using the new directory structure
func (m *Manager) generateCachePath(source, org, pluginName, version, platform string) string {
	// Create the directory structure: .dehydrated-api-go/plugins/<source>/<org>/<plugin>/<version>/<platform>/
	pluginDir := filepath.Join(m.cacheDir, source, org, pluginName, version, platform)

	// The plugin executable will be named after the plugin
	return filepath.Join(pluginDir, pluginName)
}

// downloadPlugin downloads a plugin from GitHub releases
func (m *Manager) downloadPlugin(repository, version, platform, cachePath string) (string, error) {
	// Get release information
	release, err := m.getRelease(repository, version)
	if err != nil {
		return "", err
	}

	// Find the appropriate asset for the platform
	asset, err := m.findAsset(release.Assets, platform)
	if err != nil {
		return "", err
	}

	// Create the directory for this plugin version/platform
	pluginDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Download the asset
	return m.downloadAsset(asset.BrowserDownloadURL, cachePath)
}

// getRelease fetches release information from GitHub
func (m *Manager) getRelease(repository, version string) (*GitHubRelease, error) {
	var url string
	if version == "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repository)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repository, version)
	}

	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch release info: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release info: %w", err)
	}

	return &release, nil
}

// findAsset finds the appropriate asset for the given platform
func (m *Manager) findAsset(assets []GitHubAsset, platform string) (*GitHubAsset, error) {
	// Common platform suffixes
	platformSuffixes := []string{
		platform,
		strings.ReplaceAll(platform, "-", "_"),
		strings.ReplaceAll(platform, "_", "-"),
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		for _, suffix := range platformSuffixes {
			if strings.Contains(name, strings.ToLower(suffix)) {
				return &asset, nil
			}
		}
	}

	return nil, fmt.Errorf("no asset found for platform %s", platform)
}

// downloadAsset downloads a file from the given URL
func (m *Manager) downloadAsset(url, cachePath string) (string, error) {
	m.logger.Debug("Starting download", zap.String("url", url), zap.String("cachePath", cachePath))

	resp, err := m.downloadClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset: HTTP %d", resp.StatusCode)
	}

	// Log file size if available
	if resp.ContentLength > 0 {
		m.logger.Debug("Downloading file",
			zap.String("url", url),
			zap.Int64("size", resp.ContentLength))
	}

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "dehydrated-api-plugin-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory

	// Determine if this is a compressed archive based on URL
	isArchive := strings.HasSuffix(strings.ToLower(url), ".tar.gz") ||
		strings.HasSuffix(strings.ToLower(url), ".tgz") ||
		strings.HasSuffix(strings.ToLower(url), ".zip")

	if isArchive {
		return m.handleArchiveDownload(resp.Body, tempDir, cachePath, url)
	}

	// Handle as regular file
	return m.handleRegularDownload(resp.Body, cachePath, url)
}

// handleArchiveDownload handles downloading and extracting compressed archives
func (m *Manager) handleArchiveDownload(body io.Reader, tempDir, cachePath, url string) (string, error) {
	// Create temporary archive file
	tempArchive := filepath.Join(tempDir, "archive")
	archiveFile, err := os.Create(tempArchive)
	if err != nil {
		return "", fmt.Errorf("failed to create temp archive file: %w", err)
	}
	defer archiveFile.Close()

	// Download the archive
	written, err := io.Copy(archiveFile, body)
	if err != nil {
		return "", fmt.Errorf("failed to download archive: %w", err)
	}

	m.logger.Debug("Archive download completed",
		zap.String("url", url),
		zap.Int64("bytesWritten", written))

	// Close the file before extraction
	archiveFile.Close()

	// Extract the archive to cache directory
	extractedPath, err := m.extractArchive(tempArchive, filepath.Dir(cachePath), url)
	if err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	return extractedPath, nil
}

// handleRegularDownload handles downloading regular files
func (m *Manager) handleRegularDownload(body io.Reader, cachePath, url string) (string, error) {
	// Create the file
	file, err := os.Create(cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	// Copy the response body to the file
	written, err := io.Copy(file, body)
	if err != nil {
		return "", fmt.Errorf("failed to write cache file: %w", err)
	}

	m.logger.Debug("Download completed",
		zap.String("url", url),
		zap.Int64("bytesWritten", written))

	return cachePath, nil
}

// extractArchive extracts a tar.gz or zip archive
func (m *Manager) extractArchive(archivePath, extractDir, url string) (string, error) {
	if strings.HasSuffix(strings.ToLower(url), ".zip") {
		return m.extractZip(archivePath, extractDir)
	}
	return m.extractTarGz(archivePath, extractDir)
}

// extractTarGz extracts a tar.gz archive
func (m *Manager) extractTarGz(archivePath, extractDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var mainExecutable string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Extract all regular files
		if header.Typeflag == tar.TypeReg {
			targetPath := filepath.Join(extractDir, filepath.Base(header.Name))

			// Create the file
			outFile, err := os.Create(targetPath)
			if err != nil {
				return "", fmt.Errorf("failed to create extracted file: %w", err)
			}

			// Copy the file content
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

			// Make it executable
			if err := os.Chmod(targetPath, 0755); err != nil {
				return "", fmt.Errorf("failed to make file executable: %w", err)
			}

			// Select the main executable file with the specific prefix
			fileName := filepath.Base(header.Name)
			if strings.HasPrefix(fileName, "dehydrated-api-metadata-plugin-") {
				mainExecutable = targetPath
			} else if mainExecutable == "" {
				// Fallback to first file if no prefixed file is found
				mainExecutable = targetPath
			}
		}
	}

	if mainExecutable == "" {
		return "", fmt.Errorf("no files found in archive")
	}

	return mainExecutable, nil
}

// extractZip extracts a zip archive
func (m *Manager) extractZip(archivePath, extractDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer reader.Close()

	var mainExecutable string

	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Extract all files
		targetPath := filepath.Join(extractDir, filepath.Base(file.Name))

		// Create the file
		outFile, err := os.Create(targetPath)
		if err != nil {
			return "", fmt.Errorf("failed to create extracted file: %w", err)
		}

		// Open the file in the archive
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return "", fmt.Errorf("failed to open file in archive: %w", err)
		}

		// Copy the file content
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", fmt.Errorf("failed to extract file: %w", err)
		}
		rc.Close()
		outFile.Close()

		// Make it executable
		if err := os.Chmod(targetPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make file executable: %w", err)
		}

		// Select the main executable file with the specific prefix
		fileName := filepath.Base(file.Name)
		if strings.HasPrefix(fileName, "dehydrated-api-metadata-plugin-") {
			mainExecutable = targetPath
		} else if mainExecutable == "" {
			// Fallback to first file if no prefixed file is found
			mainExecutable = targetPath
		}
	}

	if mainExecutable == "" {
		return "", fmt.Errorf("no files found in archive")
	}

	return mainExecutable, nil
}

// detectPlatform detects the current platform
func (m *Manager) detectPlatform() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go's GOOS/GOARCH to common platform names
	osMap := map[string]string{
		"linux":   "linux",
		"darwin":  "darwin",
		"windows": "windows",
		"freebsd": "freebsd",
	}

	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
		"arm":   "arm",
	}

	osName, ok := osMap[goos]
	if !ok {
		osName = goos
	}

	archName, ok := archMap[goarch]
	if !ok {
		archName = goarch
	}

	return fmt.Sprintf("%s-%s", osName, archName)
}

// isPluginCached checks if a plugin is already cached
func (m *Manager) isPluginCached(cachePath string) bool {
	info, err := os.Stat(cachePath)
	if err != nil {
		return false
	}

	// Check if it's executable
	return info.Mode()&0111 != 0
}

// Cleanup removes old cached plugins
func (m *Manager) Cleanup(maxAge time.Duration) error {
	return m.cleanupDirectory(m.cacheDir, maxAge)
}

// cleanupDirectory recursively cleans up old files in a directory
func (m *Manager) cleanupDirectory(dir string, maxAge time.Duration) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist
		}
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively clean up subdirectories
			if err := m.cleanupDirectory(entryPath, maxAge); err != nil {
				m.logger.Warn("Failed to clean up subdirectory",
					zap.String("path", entryPath),
					zap.Error(err))
			}

			// Check if directory is now empty and remove it
			if isEmpty, _ := m.isDirectoryEmpty(entryPath); isEmpty {
				if err := os.Remove(entryPath); err != nil {
					m.logger.Warn("Failed to remove empty directory",
						zap.String("path", entryPath),
						zap.Error(err))
				} else {
					m.logger.Debug("Removed empty directory", zap.String("path", entryPath))
				}
			}
		} else {
			// Check file modification time
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				if err := os.Remove(entryPath); err != nil {
					m.logger.Warn("Failed to remove old cached plugin",
						zap.String("path", entryPath),
						zap.Error(err))
				} else {
					m.logger.Debug("Removed old cached plugin", zap.String("path", entryPath))
				}
			}
		}
	}

	return nil
}

// isDirectoryEmpty checks if a directory is empty
func (m *Manager) isDirectoryEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.cacheDir
}

// GenerateCachePath generates the cache path using the new directory structure (public method)
func (m *Manager) GenerateCachePath(source, org, pluginName, version, platform string) string {
	return m.generateCachePath(source, org, pluginName, version, platform)
}
