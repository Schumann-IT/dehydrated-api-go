package registry

import (
	"github.com/schumann-it/dehydrated-api-go/dehydrated/model"
	"github.com/schumann-it/dehydrated-api-go/dehydrated/service"
	"testing"

	plugininterface "github.com/schumann-it/dehydrated-api-go/dehydrated/plugin/interface"
)

func TestRegistry(t *testing.T) {
	cfg := service.NewConfig()
	registry := NewRegistry(cfg)

	t.Run("RegisterPlugin", func(t *testing.T) {
		plugin := plugininterface.NewMockPlugin("test")
		if err := registry.Register(plugin); err != nil {
			t.Errorf("Failed to register plugin: %v", err)
		}
		if plugin.GetInitCalls() != 1 {
			t.Errorf("Expected 1 init call, got %d", plugin.GetInitCalls())
		}
	})

	t.Run("RegisterDuplicatePlugin", func(t *testing.T) {
		plugin := plugininterface.NewMockPlugin("test")
		if err := registry.Register(plugin); err == nil {
			t.Error("Expected error when registering duplicate plugin")
		}
	})

	t.Run("RegisterErrorPlugin", func(t *testing.T) {
		plugin := plugininterface.NewMockPlugin("error").WithError()
		if err := registry.Register(plugin); err == nil {
			t.Error("Expected error when registering error plugin")
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
		plugin := plugininterface.NewMockPlugin("error").WithError()
		if err := registry.Register(plugin); err == nil {
			t.Error("Expected error when registering error plugin")
		}
	})

	t.Run("Close", func(t *testing.T) {
		plugin := plugininterface.NewMockPlugin("close")
		if err := registry.Register(plugin); err != nil {
			t.Errorf("Failed to register plugin: %v", err)
		}
		if err := registry.Close(); err != nil {
			t.Errorf("Failed to close registry: %v", err)
		}
		if plugin.GetCloseCalls() != 1 {
			t.Errorf("Expected 1 close call, got %d", plugin.GetCloseCalls())
		}
	})

	t.Run("GetPlugin", func(t *testing.T) {
		plugin := plugininterface.NewMockPlugin("get")
		if err := registry.Register(plugin); err != nil {
			t.Errorf("Failed to register plugin: %v", err)
		}
		p := registry.GetPlugin("get")
		if p == nil {
			t.Error("Expected plugin to be found")
		}
		if p.Name() != "get" {
			t.Errorf("Expected plugin name 'get', got %s", p.Name())
		}
	})

	t.Run("ListPlugins", func(t *testing.T) {
		names := registry.ListPlugins()
		if len(names) == 0 {
			t.Error("Expected at least one plugin")
		}
		found := false
		for _, name := range names {
			if name == "get" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'get' plugin to be in list")
		}
	})
}

func TestRegistryConcurrency(t *testing.T) {
	cfg := service.NewConfig()
	registry := NewRegistry(cfg)

	// Create a plugin that will be registered multiple times
	plugin := plugininterface.NewMockPlugin("concurrent")

	// Test concurrent registration
	t.Run("ConcurrentRegistration", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				err := registry.Register(plugin)
				if err != nil {
					// Only the first registration should succeed
					if i == 0 && err != nil {
						t.Errorf("First registration failed: %v", err)
					}
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify only one registration succeeded
		names := registry.ListPlugins()
		if len(names) != 1 {
			t.Errorf("Expected 1 plugin, got %d", len(names))
		}
	})

	// Test concurrent enrichment
	t.Run("ConcurrentEnrichment", func(t *testing.T) {
		entry := &model.DomainEntry{
			Domain: "example.com",
		}

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				err := registry.EnrichDomainEntry(entry)
				if err != nil {
					t.Errorf("Enrichment failed: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify metadata was set correctly
		if entry.Metadata["concurrent"] != "test" {
			t.Error("Expected metadata to be set")
		}
	})

	// Test concurrent close
	t.Run("ConcurrentClose", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				err := registry.Close()
				if err != nil {
					t.Errorf("Close failed: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify registry is empty
		names := registry.ListPlugins()
		if len(names) != 0 {
			t.Errorf("Expected 0 plugins, got %d", len(names))
		}
	})
}
