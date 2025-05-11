package proto

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// Metadata represents a map of metadata values that can be converted to and from proto values
type Metadata struct {
	values map[string]interface{}
	error  string
}

// NewMetadata creates a new Metadata
func NewMetadata() *Metadata {
	return &Metadata{
		values: make(map[string]interface{}),
	}
}

// FromProto sets values from a proto value map
func (mm *Metadata) FromProto(name string, m map[string]*structpb.Value) {
	result := make(map[string]interface{})
	for k, v := range m {
		if v != nil {
			result[k] = v.AsInterface()
		}
	}
	mm.values[name] = result
}

// ToProto converts the Metadata to a proto value map
func (mm *Metadata) ToProto() (map[string]*structpb.Value, error) {
	result := make(map[string]*structpb.Value)
	for k, v := range mm.values {
		protoVal, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", k, err)
		}
		result[k] = protoVal
	}
	return result, nil
}

// SetError sets an error message for the metadata map
func (mm *Metadata) SetError(err string) {
	mm.error = err
}

// GetError returns the error message
func (mm *Metadata) GetError() string {
	return mm.error
}

// Set sets a value for the given key
func (mm *Metadata) Set(key string, value interface{}) {
	mm.values[key] = value
}

// Get returns a value for the given key
func (mm *Metadata) Get(key string) interface{} {
	return mm.values[key]
}

// ToGetMetadataResponse converts the Metadata to a GetMetadataResponse
func (mm *Metadata) ToGetMetadataResponse() (*GetMetadataResponse, error) {
	protoMap, err := mm.ToProto()
	if err != nil {
		return nil, fmt.Errorf("failed to convert metadata to proto: %w", err)
	}

	return &GetMetadataResponse{
		Metadata: protoMap,
		Error:    mm.error,
	}, nil
}
