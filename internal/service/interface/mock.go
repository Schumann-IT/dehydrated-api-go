package serviceinterface

import (
	"fmt"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// MockErrDomainService implements the DomainService interface for testing.
// It provides a simple in-memory implementation of domain operations.
type MockErrDomainService struct{}

// ListDomains returns an empty list of domains for testing.
func (m *MockErrDomainService) ListDomains() ([]*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// GetDomain returns a mock domain entry for testing.
func (m *MockErrDomainService) GetDomain(domain string) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// GetDomainByAlias returns a mock domain entry for testing.
func (m *MockErrDomainService) GetDomainByAlias(domain string, alias string) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// CreateDomain creates a mock domain entry for testing.
func (m *MockErrDomainService) CreateDomain(req model.CreateDomainRequest) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// UpdateDomain updates a mock domain entry for testing.
func (m *MockErrDomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// UpdateDomainByAlias updates a mock domain entry for testing.
func (m *MockErrDomainService) UpdateDomainByAlias(domain string, alias string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// DeleteDomain simulates deleting a domain entry for testing.
func (m *MockErrDomainService) DeleteDomain(domain string) error {
	return fmt.Errorf("mock error")
}

// DeleteDomainByAlias simulates deleting a domain entry for testing.
func (m *MockErrDomainService) DeleteDomainByAlias(domain string, alias string) error {
	return fmt.Errorf("mock error")
}

// Close performs cleanup for the mock service.
func (m *MockErrDomainService) Close() error {
	return nil
}
