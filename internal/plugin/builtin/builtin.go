package builtin

import (
	"fmt"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/builtin/openssl"
	plugininterface "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

func LoadPlugin(name string) (plugininterface.Plugin, error) {
	var server pb.PluginServer

	switch name {
	case "openssl":
		server = openssl.New()
	default:
		return nil, fmt.Errorf("built-in plugin %s not found", name)
	}

	// Create a wrapper for the built-in plugin
	return &wrapper{
		server: server,
	}, nil
}
