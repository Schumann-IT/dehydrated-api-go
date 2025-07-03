package cache

import (
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/cacheinterface"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/github"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/localfile"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
)

var (
	localCache  cacheinterface.PluginCache
	githubCache cacheinterface.PluginCache
)

func Prepare(basePath string) {
	var err error
	if basePath == "" {
		// Use current working directory for cache
		basePath, err = os.Getwd()
		if err != nil {
			// Fallback to temp directory if we can't get working directory
			basePath = filepath.Join(os.TempDir(), "dehydrated-api-go")
		}
	}

	basePath, err = filepath.Abs(filepath.Join(basePath, ".dehydrated-api-go", "plugins"))
	if err != nil {
		panic("Failed to resolve absolute path for plugin cache: " + err.Error())
	}

	if err = os.MkdirAll(basePath, 0755); err != nil {
		panic("Failed to create plugin cache directory: " + err.Error())
	}

	localCache = localfile.New(basePath)
	githubCache = github.New(basePath)
}

func Add(name string, sourceRegistry *config.RegistryConfig) cacheinterface.PluginCache {
	var c cacheinterface.PluginCache
	switch sourceRegistry.Type {
	case config.PluginSourceTypeLocal:
		c = localCache
	case config.PluginSourceTypeGitHub:
		c = githubCache
	default:
		panic("unsupported registry type: " + sourceRegistry.Type)
	}

	c.Add(name, sourceRegistry.Config)

	return c
}

func Get(name string) (string, error) {
	if localCache == nil || githubCache == nil {
		panic("plugin cache is not initialized, please call NewCache first")
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

	panic("plugin not found in any cache: " + name)
}
