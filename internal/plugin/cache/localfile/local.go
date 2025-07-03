package localfile

import (
	"encoding/json"
	"errors"
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
		path: basePath,
	}
}

func (c *LocalCache) Add(name string, s any) {
	if c.files == nil {
		c.files = map[string]string{}
	} else if _, exists := c.files[name]; exists {
		return
	}

	b, err := json.Marshal(s)
	if err != nil {
		panic("failed to marshal local cache config: " + err.Error())
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic("failed to unmarshal local cache config: " + err.Error())
	}

	i, err := os.Stat(cfg.Path)
	if err != nil || i.IsDir() {
		panic("cannot find source artifact: " + cfg.Path)
	}
	sio, err := os.OpenFile(cfg.Path, os.O_RDONLY, 0644)
	if err != nil {
		panic("cannot read source: " + cfg.Path + ": " + err.Error())
	}

	path := filepath.Join(c.path, name)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		panic("cannot create directory for target: " + path + " from source: " + cfg.Path + ": " + err.Error())
	}

	path = filepath.Join(path, i.Name())
	tio, err := os.Create(path)
	if err != nil {
		panic("cannot create target: " + path + " from source: " + cfg.Path + ": " + err.Error())
	}
	err = tio.Chmod(0766)
	if err != nil {
		panic("cannot set permissions for target: " + path + " from source: " + cfg.Path + ": " + err.Error())
	}

	_, err = io.Copy(tio, sio)
	if err != nil {
		panic("cannot create target: " + path + " from source: " + cfg.Path + ": " + err.Error())
	}

	c.files[name] = path
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
