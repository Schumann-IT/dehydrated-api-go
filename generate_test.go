//go:build testplugin
// +build testplugin

package main

//go:generate echo "Building test plugin..."
//go:generate go build -o internal/plugin/grpc/testdata/test-plugin/test-plugin internal/plugin/grpc/testdata/test-plugin/main.go
//go:generate echo "test plugin generation complete!"
