//go:build generate
// +build generate

package main

//go:generate echo "Generating protobuf files..."
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/plugin/plugin.proto
//go:generate echo "Generation complete!"
