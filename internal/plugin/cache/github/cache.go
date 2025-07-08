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
		path:  filepath.Join(basePath, "github"),
		files: map[string]string{},
		client: &http.Client{
			Timeout: 30 * time.Second, // Shorter timeout for API calls
		},
		downloadClient: &http.Client{
			Timeout: 10 * time.Minute, // Longer timeout for file downloads
		},
	}
}

func (c *GithubCache) Add(name string, s any) (cacheinterface.PluginCache, error) {
	if p, ok := c.files[name]; ok {
		_, err := os.Stat(p)
		if err == nil {
			// plugin exists
			return c, nil
		}
	}

	b, err := json.Marshal(s)
	if err != nil {
		return c, fmt.Errorf("error marshaling %v: %w", name, err)
	}
	var gcfg GitHubConfig
	err = json.Unmarshal(b, &gcfg)
	if err != nil {
		return c, fmt.Errorf("error unmarshalling %v: %w", name, err)
	}

	resp, err := c.client.Get(gcfg.getReleaseUrl())
	if err != nil {
		return c, fmt.Errorf("error fetching release info %v: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c, fmt.Errorf("error fetching release info %s:%v: %v", gcfg.getReleaseUrl(), name, resp.Status)
	}

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return c, fmt.Errorf("error decoding release info %v: %w", name, err)
	}
	asset, err := findAsset(release.Assets, gcfg.getPlatform())
	if err != nil {
		return c, fmt.Errorf("failed to find asset for platform %s %v: %w", gcfg.Platform, name, err)
	}

	c.currentFile = filepath.Join(c.path, gcfg.getOrg(), gcfg.getName(), name, gcfg.getVersion(), gcfg.getPlatform(), gcfg.getName())
	err = os.MkdirAll(filepath.Dir(c.currentFile), 0755)
	if err != nil {
		return c, fmt.Errorf("error creating target directory %v: %w", filepath.Dir(c.currentFile), err)
	}

	err = c.downloadAsset(asset)
	if err != nil {
		return c, fmt.Errorf("error downloading asset %s: %w", asset.BrowserDownloadURL, err)
	}

	c.files[name] = c.currentFile

	return c, nil
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
func (c *GithubCache) downloadAsset(asset *GitHubAsset) error {
	resp, err := c.downloadClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: HTTP %v for %s", resp.Status, asset.BrowserDownloadURL)
	}

	return c.doDownload(asset, resp)
}

func (c *GithubCache) doDownload(asset *GitHubAsset, resp *http.Response) error {
	// Determine if this is a compressed archive based on URL
	isArchive := strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tar.gz") ||
		strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tgz") ||
		strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".zip")

	if isArchive {
		archiveFile, tmpFile, err := handleArchiveDownload(resp)
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpFile)

		if strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".zip") {
			return extractZip(archiveFile, filepath.Dir(c.currentFile))
		} else if strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tar.gz") ||
			strings.HasSuffix(strings.ToLower(asset.BrowserDownloadURL), ".tgz") {
			return extractTarGz(archiveFile, filepath.Dir(c.currentFile))
		}
	}

	// If not an archive, handle as a regular file
	return c.handleRegularDownload(resp)
}

// handleArchiveDownload handles downloading and extracting compressed archives
func handleArchiveDownload(resp *http.Response) (string, string, error) {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "dehydrated-api-plugin-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory for archive: %w", err)
	}

	// Create temporary archive file
	tempArchive := filepath.Join(tempDir, "archive")
	archiveFile, err := os.Create(tempArchive)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary archive file: %w", err)
	}
	defer archiveFile.Close()

	// Download the archive
	_, err = io.Copy(archiveFile, resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to copy response body to archive file: %w", err)
	}

	return tempArchive, tempDir, nil
}

// handleRegularDownload handles downloading regular files
func (c *GithubCache) handleRegularDownload(resp *http.Response) error {
	// Create the file
	file, err := os.Create(c.currentFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// extractTarGz extracts a tar.gz archive
func extractTarGz(archivePath, extractDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	mainExecutable, err := doExtractTarGz(gzr, extractDir)
	if err != nil {
		return err
	}
	if mainExecutable == "" {
		return errors.New("no files found in archive")
	}

	// Make it executable
	return os.Chmod(mainExecutable, 0755)
}

func doExtractTarGz(gzr *gzip.Reader, extractDir string) (string, error) {
	tr := tar.NewReader(gzr)
	var mainExecutable string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
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
				return "", err
			}

			// Copy the file content
			//nolint:gosec // We trust the source of the archive
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
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

	return mainExecutable, nil
}

// extractZip extracts a zip archive
func extractZip(archivePath, extractDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	mainExecutable, err := doExtractZip(reader, extractDir)
	if err != nil {
		return err
	}

	if mainExecutable == "" {
		return errors.New("no files found in archive")
	}

	// Make it executable
	return os.Chmod(mainExecutable, 0755)
}

func doExtractZip(reader *zip.ReadCloser, extractDir string) (string, error) {
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
			return "", err
		}

		// Open the file in the archive
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		// Copy the file content
		//nolint:gosec // We trust the source of the archive
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", err
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
	return mainExecutable, nil
}

func (c *GithubCache) Clean() {
	_ = os.RemoveAll(c.path)
}
