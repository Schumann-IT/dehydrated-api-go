package plugin

import (
	"context"
	"testing"
)

// MockPlugin implements the Plugin interface for testing
type MockPlugin struct {
	name         string
	initCalls    int
	enrichCalls  int
	closeCalls   int
	shouldError  bool
	initConfig   map[string]any
	enrichDomain *Domain
	enrichResult *Metadata
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

// GetEnrichCalls returns the number of EnrichDomainEntry calls
func (p *MockPlugin) GetEnrichCalls() int {
	return p.enrichCalls
}

// GetCloseCalls returns the number of Close calls
func (p *MockPlugin) GetCloseCalls() int {
	return p.closeCalls
}

// GetInitConfig returns the last config passed to Initialize
func (p *MockPlugin) GetInitConfig() map[string]any {
	return p.initConfig
}

// GetEnrichDomain returns the last domain passed to EnrichDomainEntry
func (p *MockPlugin) GetEnrichDomain() *Domain {
	return p.enrichDomain
}

// GetEnrichResult returns the last result from EnrichDomainEntry
func (p *MockPlugin) GetEnrichResult() *Metadata {
	return p.enrichResult
}

func (p *MockPlugin) Initialize(ctx context.Context, config map[string]any) error {
	p.initCalls++
	p.initConfig = config
	if p.shouldError {
		return ErrPluginError
	}
	return nil
}

func (p *MockPlugin) EnrichDomainEntry(ctx context.Context, domain *Domain) (*Metadata, error) {
	p.enrichCalls++
	p.enrichDomain = domain
	if p.shouldError {
		return nil, ErrPluginError
	}
	result := NewMetadata()
	result.Values["test"] = "value"
	p.enrichResult = result
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
	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
	if plugin.GetInitCalls() != 1 {
		t.Errorf("Expected 1 init call, got %d", plugin.GetInitCalls())
	}
	if plugin.GetInitConfig()["key"] != "value" {
		t.Errorf("Expected config key=value, got %v", plugin.GetInitConfig())
	}

	// Test EnrichDomainEntry
	domain := NewDomain("example.com")
	metadata, err := plugin.EnrichDomainEntry(ctx, domain)
	if err != nil {
		t.Errorf("EnrichDomainEntry failed: %v", err)
	}
	if plugin.GetEnrichCalls() != 1 {
		t.Errorf("Expected 1 enrich call, got %d", plugin.GetEnrichCalls())
	}
	if plugin.GetEnrichDomain().Name != "example.com" {
		t.Errorf("Expected domain example.com, got %s", plugin.GetEnrichDomain().Name)
	}
	if metadata.Values["test"] != "value" {
		t.Errorf("Expected metadata test=value, got %v", metadata.Values)
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
	if err := plugin.Initialize(ctx, config); err != ErrPluginError {
		t.Errorf("Expected Initialize error, got %v", err)
	}
	if metadata, err := plugin.EnrichDomainEntry(ctx, domain); err != ErrPluginError {
		t.Errorf("Expected EnrichDomainEntry error, got %v", err)
	} else if metadata != nil {
		t.Error("Expected nil metadata on error")
	}
	if err := plugin.Close(ctx); err != ErrPluginError {
		t.Errorf("Expected Close error, got %v", err)
	}
}
