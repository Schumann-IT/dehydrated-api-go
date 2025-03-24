package plugininterface

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// ConvertToStructValue converts a map[string]any to map[string]*structpb.Value
// This is a utility function used by plugin implementations to convert Go types to protobuf types.
func ConvertToStructValue(config map[string]any) (map[string]*structpb.Value, error) {
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
