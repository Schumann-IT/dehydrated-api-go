package cacheinterface

type PluginCache interface {
	Add(string, any) (PluginCache, error)
	Path(string) (string, error)
	Clean()
}
