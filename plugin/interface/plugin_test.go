package plugininterface

import (
	"context"
	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated/model"
	"testing"
)

// MockPlugin implements the Plugin interface for testing
type MockPlugin struct {
	name                string
	initCalls           int
	metadataCalls       int
	closeCalls          int
	shouldError         bool
	initConfig          map[string]any
	metadataDomainEntry string
	metadataResult      map[string]any
}

// NewMockPlugin creates a new mock plugin
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name: name,
	}
}

// WithError makes the plugin return errors
func (p *MockPlugin) WithError() *MockPlugin {
	p.shouldError = true
	return p
}

// GetInitCalls returns the number of Initialize calls
func (p *MockPlugin) GetInitCalls() int {
	return p.initCalls
}

// GetMetadataCalls returns the number of GetMetadata calls
func (p *MockPlugin) GetMetadataCalls() int {
	return p.metadataCalls
}

// GetCloseCalls returns the number of Close calls
func (p *MockPlugin) GetCloseCalls() int {
	return p.closeCalls
}

// GetInitConfig returns the last config passed to Initialize
func (p *MockPlugin) GetInitConfig() map[string]any {
	return p.initConfig
}

// GetMetadataDomain returns the last domain passed to GetMetadata
func (p *MockPlugin) GetMetadataDomain() string {
	return p.metadataDomainEntry
}

// GetMetadataResult returns the last result from GetMetadata
func (p *MockPlugin) GetMetadataResult() map[string]any {
	return p.metadataResult
}

func (p *MockPlugin) Initialize(config map[string]any) error {
	p.initCalls++
	p.initConfig = config
	if p.shouldError {
		return ErrPluginError
	}
	return nil
}

func (p *MockPlugin) GetMetadata(entry model.DomainEntry) (map[string]any, error) {
	p.metadataCalls++
	p.metadataDomainEntry = entry.Domain
	if p.shouldError {
		return nil, ErrPluginError
	}
	result := map[string]any{"test": "value"}
	p.metadataResult = result
	return result, nil
}

func (p *MockPlugin) Close(ctx context.Context) error {
	p.closeCalls++
	if p.shouldError {
		return ErrPluginError
	}
	return nil
}

func TestMockPlugin(t *testing.T) {
	ctx := context.Background()
	plugin := NewMockPlugin("test")

	// Test Initialize
	config := map[string]any{"key": "value"}
	err := plugin.Initialize(config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
	if plugin.GetInitCalls() != 1 {
		t.Errorf("Expected 1 init call, got %d", plugin.GetInitCalls())
	}
	if plugin.GetInitConfig()["key"] != "value" {
		t.Errorf("Expected config key=value, got %v", plugin.GetInitConfig())
	}

	// Test GetMetadata
	domain := model.DomainEntry{
		Domain: "example.com",
	}
	metadata, err := plugin.GetMetadata(domain)
	if err != nil {
		t.Errorf("GetMetadata failed: %v", err)
	}
	if plugin.GetMetadataCalls() != 1 {
		t.Errorf("Expected 1 metadata call, got %d", plugin.GetMetadataCalls())
	}
	if plugin.GetMetadataDomain() != "example.com" {
		t.Errorf("Expected domain example.com, got %s", plugin.GetMetadataDomain())
	}
	if metadata["test"] != "value" {
		t.Errorf("Expected metadata test=value, got %v", metadata)
	}

	// Test Close
	err = plugin.Close(ctx)
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if plugin.GetCloseCalls() != 1 {
		t.Errorf("Expected 1 close call, got %d", plugin.GetCloseCalls())
	}

	// Test error handling
	plugin.WithError()
	if err := plugin.Initialize(config); err != ErrPluginError {
		t.Errorf("Expected Initialize error, got %v", err)
	}
	if metadata, err := plugin.GetMetadata(domain); err != ErrPluginError {
		t.Errorf("Expected GetMetadata error, got %v", err)
	} else if metadata != nil {
		t.Error("Expected nil metadata on error")
	}
	if err := plugin.Close(ctx); err != ErrPluginError {
		t.Errorf("Expected Close error, got %v", err)
	}
}
