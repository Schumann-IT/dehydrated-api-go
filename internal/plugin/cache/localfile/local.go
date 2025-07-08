package localfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache/cacheinterface"
)

type LocalCache struct {
	path  string
	files map[string]string
}

func New(basePath string) cacheinterface.PluginCache {
	return &LocalCache{
		path:  filepath.Join(basePath, "local"),
		files: map[string]string{},
	}
}

func (c *LocalCache) Add(name string, s any) (cacheinterface.PluginCache, error) {
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
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return c, fmt.Errorf("error unmarshalling %v: %w", name, err)
	}

	i, err := os.Stat(cfg.Path)
	if err != nil || i.IsDir() {
		return c, fmt.Errorf("%v is not a file", cfg.Path)
	}
	sio, err := os.OpenFile(cfg.Path, os.O_RDONLY, 0644)
	if err != nil {
		return c, fmt.Errorf("error opening source file %v: %w", cfg.Path, err)
	}

	path := filepath.Join(c.path, name)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return c, fmt.Errorf("error creating target directory %v: %w", path, err)
	}

	path = filepath.Join(path, i.Name())
	tio, err := os.Create(path)
	if err != nil {
		return c, fmt.Errorf("error creating target %s: %w", path, err)
	}
	err = tio.Chmod(0766)
	if err != nil {
		return c, fmt.Errorf("cannot set permissions for target %s: %w", path, err)
	}

	_, err = io.Copy(tio, sio)
	if err != nil {
		return c, fmt.Errorf("error writing target %s: %w", path, err)
	}

	c.files[name] = path

	return c, nil
}

func (c *LocalCache) Path(name string) (string, error) {
	if c.files == nil {
		return "", errors.New("cache is empty")
	}
	if _, exists := c.files[name]; !exists {
		return "", errors.New("plugin " + name + " not found")
	}

	return c.files[name], nil
}

func (c *LocalCache) Clean() {
	_ = os.RemoveAll(c.path)
}
