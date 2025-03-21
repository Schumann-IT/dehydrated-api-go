package plugininterface

import (
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal/config"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// MockPlugin is a test implementation of the Plugin interface
type MockPlugin struct {
	name        string
	initCalls   int
	enrichCalls int
	closeCalls  int
	shouldError bool
	mu          sync.Mutex
}

// NewMockPlugin creates a new mock plugin with the given name
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{name: name}
}

// WithError makes the mock plugin return errors
func (p *MockPlugin) WithError() *MockPlugin {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shouldError = true
	return p
}

// GetInitCalls returns the number of times Initialize was called
func (p *MockPlugin) GetInitCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.initCalls
}

// GetEnrichCalls returns the number of times EnrichDomainEntry was called
func (p *MockPlugin) GetEnrichCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enrichCalls
}

// GetCloseCalls returns the number of times Close was called
func (p *MockPlugin) GetCloseCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeCalls
}

func (p *MockPlugin) Name() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.name
}

func (p *MockPlugin) Initialize(cfg *config.Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initCalls++
	if p.shouldError {
		return &PluginError{
			Name:    p.name,
			Message: "initialization failed",
		}
	}
	return nil
}

func (p *MockPlugin) EnrichDomainEntry(entry *model.DomainEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enrichCalls++
	if p.shouldError {
		return &PluginError{
			Name:    p.name,
			Message: "enrichment failed",
		}
	}
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	entry.Metadata[p.name] = "test"
	return nil
}

func (p *MockPlugin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeCalls++
	if p.shouldError {
		return &PluginError{
			Name:    p.name,
			Message: "close failed",
		}
	}
	return nil
}
