package github

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/cacheinterface"
)

type GithubCache struct {
	path           string
	files          map[string]string
	client         *http.Client
	downloadClient *http.Client
	currentFile    string
}

func New(basePath string) cacheinterface.PluginCache {
	return &GithubCache{
		path: filepath.Join(basePath, "github"),
		client: &http.Client{
			Timeout: 30 * time.Second, // Shorter timeout for API calls
		},
		downloadClient: &http.Client{
			Timeout: 10 * time.Minute, // Longer timeout for file downloads
		},
	}
}

func (c *GithubCache) Add(name string, s any) {
	if c.files == nil {
		c.files = map[string]string{}
	} else if _, exists := c.files[name]; exists {
		return
	}

	b, err := json.Marshal(s)
	if err != nil {
		panic("failed to marshal githubCache config: " + err.Error())
	}
	var gcfg GitHubConfig
	err = json.Unmarshal(b, &gcfg)
	if err != nil {
		panic("failed to unmarshal githubCache config: " + err.Error())
	}

	resp, err := c.client.Get(gcfg.getReleaseUrl())
	if err != nil {
		panic("failed to fetch release info: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("failed to fetch release info: HTTP " + resp.Status + " for " + gcfg.getReleaseUrl())
	}

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		panic("failed to decode release info: " + err.Error())
	}
	asset, err := findAsset(release.Assets, gcfg.getPlatform())
	if err != nil {
		panic("failed to find asset for platform " + gcfg.getPlatform() + ": " + err.Error())
	}

	c.currentFile = filepath.Join(c.path, gcfg.getOrg(), gcfg.getName(), name, gcfg.getVersion(), gcfg.getPlatform(), gcfg.getName())
	if _, err = os.Stat(c.currentFile); err == nil {
		c.files[name] = c.currentFile
		return
	}

	err = os.MkdirAll(filepath.Dir(c.currentFile), 0755)
	if err != nil {
		panic("failed to create plugin directory: " + err.Error())
	}

	c.downloadAsset(asset)

	c.files[name] = c.currentFile
}

func (c *GithubCache) Path(name string) (string, error) {
	if c.files == nil {
		return "", errors.New("cache is empty")
	}
	if _, exists := c.files[name]; !exists {
		return "", errors.New("plugin " + name + " not found")
	}

	return c.files[name], nil
}

// findAsset finds the appropriate asset for the given platform
func findAsset(assets []GitHubAsset, platform string) (*GitHubAsset, error) {
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
func (c *GithubCache) downloadAsset(asset *GitHubAsset) {
	resp, err := c.downloadClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		panic("failed to download asset: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("failed to download asset: HTTP " + resp.Status + " for " + asset.BrowserDownloadURL)
	}

	// Determine if this is a compressed archive based on URL
	isArchive := strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tar.gz") ||
		strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tgz") ||
		strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".zip")

	if isArchive {
		archiveFile, tmpFile := handleArchiveDownload(resp)
		if strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".zip") {
			extractZip(archiveFile, filepath.Dir(c.currentFile))
		} else if strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tar.gz") ||
			strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tgz") {
			extractTarGz(archiveFile, filepath.Dir(c.currentFile))
		}
		os.RemoveAll(tmpFile)
	} else {
		// If not an archive, handle as a regular file
		c.handleRegularDownload(resp)
	}
}

// handleArchiveDownload handles downloading and extracting compressed archives
func handleArchiveDownload(resp *http.Response) (string, string) {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "dehydrated-api-plugin-*")
	if err != nil {
		panic("failed to create temp directory for archive: " + err.Error())
	}

	// Create temporary archive file
	tempArchive := filepath.Join(tempDir, "archive")
	archiveFile, err := os.Create(tempArchive)
	if err != nil {
		panic("failed to create temporary archive file: " + err.Error())
	}

	// Download the archive
	_, err = io.Copy(archiveFile, resp.Body)
	if err != nil {
		panic("failed to write archive file: " + err.Error())
	}

	archiveFile.Close()

	return tempArchive, tempDir
}

// handleRegularDownload handles downloading regular files
func (c *GithubCache) handleRegularDownload(resp *http.Response) {
	// Create the file
	file, err := os.Create(c.currentFile)
	if err != nil {
		panic("failed to create cache file: " + err.Error())
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		panic("failed to write cache file: " + err.Error())
	}
}

// extractTarGz extracts a tar.gz archive
func extractTarGz(archivePath, extractDir string) {
	file, err := os.Open(archivePath)
	if err != nil {
		panic("failed to open tar.gz archive: " + err.Error())
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		panic("failed to create gzip reader: " + err.Error())
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
			panic("failed to read tar header: " + err.Error())
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
				panic("failed to create extracted file: " + err.Error())
			}

			// Copy the file content
			//nolint:gosec // We trust the source of the archive
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				panic("failed to extract file: " + err.Error())
			}
			outFile.Close()

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
		panic("no files found in archive")
	}

	// Make it executable
	if err := os.Chmod(mainExecutable, 0755); err != nil {
		panic("failed to make file executable: " + err.Error())
	}
}

// extractZip extracts a zip archive
func extractZip(archivePath, extractDir string) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		panic("failed to open zip archive: " + err.Error())
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
			panic("failed to create extracted file: " + err.Error())
		}

		// Open the file in the archive
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			panic("failed to open file in archive: " + err.Error())
		}

		// Copy the file content
		//nolint:gosec // We trust the source of the archive
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			panic("failed to extract file: " + err.Error())
		}
		rc.Close()
		outFile.Close()

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
		panic("no files found in archive")
	}

	// Make it executable
	if err := os.Chmod(mainExecutable, 0755); err != nil {
		panic("failed to make file executable: " + err.Error())
	}
}

func (c *GithubCache) Clean() {
	_ = os.RemoveAll(c.path)
}
