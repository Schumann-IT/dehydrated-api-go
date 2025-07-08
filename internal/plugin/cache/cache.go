package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/cacheinterface"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/github"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/localfile"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
)

var (
	cacheBasePath string
	localCache    cacheinterface.PluginCache
	githubCache   cacheinterface.PluginCache
)

func Prepare(basePath string) error {
	var err error
	if basePath == "" {
		// Use current working directory for cache
		basePath, err = os.Getwd()
		if err != nil {
			// Fallback to temp directory if we can't get working directory
			basePath = filepath.Join(os.TempDir(), "dehydrated-api-go")
		}
	}

	cacheBasePath, err = filepath.Abs(filepath.Join(basePath, ".dehydrated-api-go"))
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for cache: %w", err)
	}
	basePath = filepath.Join(cacheBasePath, "plugins")
	if err = os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create plugin cache directory: %w", err)
	}

	localCache = localfile.New(basePath)
	githubCache = github.New(basePath)

	return nil
}

func Add(name string, sourceRegistry *config.RegistryConfig) (cacheinterface.PluginCache, error) {
	var c cacheinterface.PluginCache
	switch sourceRegistry.Type {
	case config.PluginSourceTypeLocal:
		c = localCache
	case config.PluginSourceTypeGitHub:
		c = githubCache
	default:
		return nil, fmt.Errorf("unsupported source type: %v", sourceRegistry.Type)
	}

	return c.Add(name, sourceRegistry.Config)
}

func Get(name string) (string, error) {
	if localCache == nil && githubCache == nil {
		return "", fmt.Errorf("plugin cache is not initialized, please call cache.Prepare() first")
	}

	// Try to get from local cache first
	path, err := localCache.Path(name)
	if err == nil && path != "" {
		return path, nil
	}

	path, err = githubCache.Path(name)
	if err == nil && path != "" {
		return path, nil
	}

	return "", fmt.Errorf("plugin %s not found in any cache", name)
}

func Clean() {
	if localCache != nil {
		localCache.Clean()
		localCache = nil
	}
	if githubCache != nil {
		githubCache.Clean()
		githubCache = nil
	}

	_ = os.RemoveAll(cacheBasePath)
}
