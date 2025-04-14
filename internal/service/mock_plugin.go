// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"context"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// MockPlugin implements the plugin interface for testing purposes.
// It allows setting custom behavior for each method through function fields.
type MockPlugin struct {
	initializeFunc  func(context.Context, map[string]interface{}) error
	getMetadataFunc func(context.Context, model.DomainEntry) (map[string]interface{}, error)
	closeFunc       func() error
}

// NewMockPlugin creates a new MockPlugin with default behavior.
// The default implementation returns empty results with no errors.
func NewMockPlugin() *MockPlugin {
	return &MockPlugin{
		initializeFunc: func(ctx context.Context, config map[string]interface{}) error {
			return nil
		},
		getMetadataFunc: func(ctx context.Context, entry model.DomainEntry) (map[string]interface{}, error) {
			return map[string]interface{}{"key": "value"}, nil
		},
		closeFunc: func() error {
			return nil
		},
	}
}

// Initialize calls the configured initialize function or the default implementation.
// It allows testing initialization behavior by setting a custom function.
func (m *MockPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	return m.initializeFunc(ctx, config)
}

// GetMetadata calls the configured getMetadata function or the default implementation.
// It allows testing metadata retrieval behavior by setting a custom function.
func (m *MockPlugin) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]interface{}, error) {
	return m.getMetadataFunc(ctx, entry)
}

// Close calls the configured close function or the default implementation.
// It allows testing cleanup behavior by setting a custom function.
func (m *MockPlugin) Close() error {
	return m.closeFunc()
}

// SetInitializeFunc sets a custom function to handle initialization.
// This is used in tests to simulate different initialization scenarios.
func (m *MockPlugin) SetInitializeFunc(f func(context.Context, map[string]interface{}) error) {
	m.initializeFunc = f
}

// SetGetMetadataFunc sets a custom function to handle metadata retrieval.
// This is used in tests to simulate different metadata scenarios.
func (m *MockPlugin) SetGetMetadataFunc(f func(context.Context, model.DomainEntry) (map[string]interface{}, error)) {
	m.getMetadataFunc = f
}

// SetCloseFunc sets a custom function to handle cleanup.
// This is used in tests to simulate different cleanup scenarios.
func (m *MockPlugin) SetCloseFunc(f func() error) {
	m.closeFunc = f
}
