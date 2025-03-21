package plugin

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"testing"
)

// MockPlugin is a test implementation of the Plugin interface
type MockPlugin struct {
	name        string
	initCalls   int
	enrichCalls int
	closeCalls  int
	shouldError bool
}

func (p *MockPlugin) Name() string {
	return p.name
}

func (p *MockPlugin) Initialize(cfg *service.Config) error {
	p.initCalls++
	if p.shouldError {
		return fmt.Errorf("mock error")
	}
	return nil
}

func (p *MockPlugin) EnrichDomainEntry(entry *model.DomainEntry) error {
	p.enrichCalls++
	if p.shouldError {
		return fmt.Errorf("mock error")
	}
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	entry.Metadata[p.name] = "test"
	return nil
}

func (p *MockPlugin) Close() error {
	p.closeCalls++
	if p.shouldError {
		return fmt.Errorf("mock error")
	}
	return nil
}

func TestPluginRegistry(t *testing.T) {
	cfg := service.NewConfig()
	registry := NewRegistry(cfg)

	t.Run("RegisterPlugin", func(t *testing.T) {
		plugin := &MockPlugin{name: "test"}
		if err := registry.Register(plugin); err != nil {
			t.Errorf("Failed to register plugin: %v", err)
		}
		if plugin.initCalls != 1 {
			t.Errorf("Expected 1 init call, got %d", plugin.initCalls)
		}
	})

	t.Run("RegisterPluginError", func(t *testing.T) {
		plugin := &MockPlugin{name: "error", shouldError: true}
		if err := registry.Register(plugin); err == nil {
			t.Error("Expected error when registering plugin")
		}
	})

	t.Run("EnrichDomainEntry", func(t *testing.T) {
		entry := &model.DomainEntry{
			Domain: "example.com",
		}
		if err := registry.EnrichDomainEntry(entry); err != nil {
			t.Errorf("Failed to enrich domain entry: %v", err)
		}
		if entry.Metadata["test"] != "test" {
			t.Error("Expected metadata to be set")
		}
	})

	t.Run("EnrichDomainEntryError", func(t *testing.T) {
		plugin := &MockPlugin{name: "error", shouldError: true}
		if err := registry.Register(plugin); err == nil {
			t.Error("Expected error when registering plugin")
		}
	})

	t.Run("Close", func(t *testing.T) {
		plugin := &MockPlugin{name: "close"}
		if err := registry.Register(plugin); err != nil {
			t.Errorf("Failed to register plugin: %v", err)
		}
		if err := registry.Close(); err != nil {
			t.Errorf("Failed to close registry: %v", err)
		}
		if plugin.closeCalls != 1 {
			t.Errorf("Expected 1 close call, got %d", plugin.closeCalls)
		}
	})
}
