package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
)

// TestDomainEntryProtoConsistency ensures that DomainEntry struct matches the GetMetadataRequest proto definition
func TestDomainEntryProtoConsistency(t *testing.T) {
	protoType := reflect.TypeOf(plugin.GetMetadataRequest{})
	modelType := reflect.TypeOf(DomainEntry{})

	// Check that all proto fields exist in model with correct types
	for i := 0; i < protoType.NumField(); i++ {
		protoField := protoType.Field(i)
		// Skip internal proto fields
		if strings.HasPrefix(protoField.Name, "XXX_") || protoField.Name == "state" || protoField.Name == "sizeCache" || protoField.Name == "unknownFields" {
			continue
		}
		modelField, exists := modelType.FieldByName(protoField.Name)
		assert.True(t, exists, "Proto field %s not found in model", protoField.Name)
		if exists {
			assert.Equal(t, protoField.Tag.Get("protobuf"), modelField.Tag.Get("protobuf"),
				"Field %s has different tags: proto=%s, model=%s",
				protoField.Name, protoField.Tag.Get("protobuf"), modelField.Tag.Get("protobuf"))
		}
	}

	// Check that all model fields exist in proto with correct types
	for i := 0; i < modelType.NumField(); i++ {
		modelField := modelType.Field(i)
		protoField, exists := protoType.FieldByName(modelField.Name)
		assert.True(t, exists, "Model field %s not found in proto", modelField.Name)
		if exists {
			assert.Equal(t, modelField.Tag.Get("protobuf"), protoField.Tag.Get("protobuf"),
				"Field %s has different tags: proto=%s, model=%s",
				modelField.Name, protoField.Tag.Get("protobuf"), modelField.Tag.Get("protobuf"))
		}
	}
}

// TestDomainEntryTypeConsistency ensures type consistency between proto and model fields
func TestDomainEntryTypeConsistency(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected reflect.Type
	}{
		{
			name:     "Domain",
			field:    "Domain",
			expected: reflect.TypeOf(""),
		},
		{
			name:     "AlternativeNames",
			field:    "AlternativeNames",
			expected: reflect.TypeOf([]string{}),
		},
		{
			name:     "Alias",
			field:    "Alias",
			expected: reflect.TypeOf(""),
		},
		{
			name:     "Enabled",
			field:    "Enabled",
			expected: reflect.TypeOf(true),
		},
		{
			name:     "Comment",
			field:    "Comment",
			expected: reflect.TypeOf(""),
		},
		{
			name:     "Metadata",
			field:    "Metadata",
			expected: reflect.TypeOf(map[string]any{}),
		},
	}

	modelType := reflect.TypeOf(DomainEntry{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := modelType.FieldByName(tt.field)
			assert.True(t, ok, "Field %s not found in DomainEntry", tt.field)
			assert.Equal(t, tt.expected, field.Type, "Field %s has wrong type", tt.field)
		})
	}
}
