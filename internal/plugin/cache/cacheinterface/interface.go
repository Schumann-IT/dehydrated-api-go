package cacheinterface

type PluginCache interface {
	Add(string, any)
	Path(string) (string, error)
}
