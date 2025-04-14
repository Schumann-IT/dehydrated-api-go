package dehydrated

import (
	"reflect"
	"testing"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

// TestDehydratedConfigProtoConsistency ensures that dehydrated.Config struct matches the DehydratedConfig proto definition
func TestDehydratedConfigProtoConsistency(t *testing.T) {
	// Get proto message type
	protoType := reflect.TypeOf((*pb.DehydratedConfig)(nil)).Elem()

	// Get internal model type
	modelType := reflect.TypeOf(Config{})

	// Compare fields
	protoFields := getProtoFields(protoType)
	modelFields := getModelFields(modelType)

	// Verify all proto fields exist in model
	for field, protoTag := range protoFields {
		if modelTag, exists := modelFields[field]; !exists {
			t.Errorf("Model missing proto field: %s", field)
		} else if modelTag != protoTag {
			t.Errorf("Field %s has different tags: proto=%s, model=%s",
				field, protoTag, modelTag)
		}
	}

	// Verify model doesn't have extra fields that should be in proto
	for field := range modelFields {
		if _, exists := protoFields[field]; !exists {
			t.Errorf("Model has extra field not in proto: %s", field)
		}
	}
}

// TestDehydratedConfigTypeConsistency ensures type consistency between proto and model fields
func TestDehydratedConfigTypeConsistency(t *testing.T) {
	protoConfig := &pb.DehydratedConfig{}
	modelConfig := &Config{}

	tests := []struct {
		name     string
		proto    interface{}
		model    interface{}
		expected bool
	}{
		{
			name:     "CertDir field",
			proto:    protoConfig.CertDir,
			model:    modelConfig.CertDir,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoType := reflect.TypeOf(tt.proto)
			modelType := reflect.TypeOf(tt.model)

			if protoType != modelType {
				t.Errorf("Type mismatch for %s: proto=%v, model=%v",
					tt.name, protoType, modelType)
			}
		})
	}
}

// getProtoFields extracts the field names and protobuf tags from a proto message type.
// It returns a map of field names to their protobuf tags.
func getProtoFields(t reflect.Type) map[string]string {
	fields := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("protobuf"); tag != "" {
			fields[field.Name] = tag
		}
	}
	return fields
}

// getModelFields extracts the field names and protobuf tags from a model type.
// It returns a map of field names to their protobuf tags.
func getModelFields(t reflect.Type) map[string]string {
	fields := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("protobuf"); tag != "" {
			fields[field.Name] = tag
		}
	}
	return fields
}
