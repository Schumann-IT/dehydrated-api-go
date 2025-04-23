// Package plugininterface provides utility functions for plugin implementations.
// It includes helper functions for type conversion and data manipulation.
package grpc

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// convertToStructValue converts a map[string]any to map[string]*structpb.Value
// This utility function is used by plugin implementations to convert Go types to protobuf types.
// It handles the conversion of various Go types to their protobuf equivalents.
// Returns an error if any value cannot be converted to a protobuf value.
func convertToStructValue(config map[string]any) (map[string]*structpb.Value, error) {
	if config == nil {
		return nil, nil
	}

	result := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", k, err)
		}
		result[k] = value
	}
	return result, nil
}
