package service

import (
	"context"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

type MockPlugin struct {
	initializeFunc  func(context.Context, map[string]interface{}) error
	getMetadataFunc func(context.Context, model.DomainEntry) (map[string]interface{}, error)
	closeFunc       func() error
}

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

func (m *MockPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	return m.initializeFunc(ctx, config)
}

func (m *MockPlugin) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]interface{}, error) {
	return m.getMetadataFunc(ctx, entry)
}

func (m *MockPlugin) Close() error {
	return m.closeFunc()
}

func (m *MockPlugin) SetInitializeFunc(f func(context.Context, map[string]interface{}) error) {
	m.initializeFunc = f
}

func (m *MockPlugin) SetGetMetadataFunc(f func(context.Context, model.DomainEntry) (map[string]interface{}, error)) {
	m.getMetadataFunc = f
}

func (m *MockPlugin) SetCloseFunc(f func() error) {
	m.closeFunc = f
}
