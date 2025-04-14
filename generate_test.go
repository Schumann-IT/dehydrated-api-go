//go:build test
// +build test

package main

//go:generate echo "Building test plugin..."
//go:generate go build -o internal/plugin/registry/testdata/test-plugin/test-plugin internal/plugin/registry/testdata/test-plugin/main.go
//go:generate echo "test plugin generation complete!"
