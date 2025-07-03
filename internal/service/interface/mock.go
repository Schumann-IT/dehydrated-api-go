//nolint: revive // This is a mock

package serviceinterface

import (
	"fmt"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
)

// MockDomainService implements the DomainService interface for testing.
// It provides a simple in-memory implementation of domain operations that returns successful responses.
type MockDomainService struct{}

// ListDomains returns an empty list of domains for testing.
func (m *MockDomainService) ListDomains() ([]*model.DomainEntry, error) {
	return []*model.DomainEntry{}, nil
}

// GetDomain returns a mock domain entry for testing.
func (m *MockDomainService) GetDomain(domain, _ string) (*model.DomainEntry, error) {
	return &model.DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:  domain,
			Enabled: true,
		},
	}, nil
}

// CreateDomain creates a mock domain entry for testing.
func (m *MockDomainService) CreateDomain(req *model.CreateDomainRequest) (*model.DomainEntry, error) {
	return &model.DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:  req.Domain,
			Enabled: req.Enabled,
		},
	}, nil
}

// UpdateDomain updates a mock domain entry for testing.
func (m *MockDomainService) UpdateDomain(domain string, _ model.UpdateDomainRequest) (*model.DomainEntry, error) {
	return &model.DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:  domain,
			Enabled: true,
		},
	}, nil
}

// DeleteDomain simulates deleting a domain entry for testing.
func (m *MockDomainService) DeleteDomain(_ string, _ model.DeleteDomainRequest) error {
	return nil
}

// Close performs cleanup for the mock service.
func (m *MockDomainService) Close() error {
	return nil
}

// MockErrDomainService implements the DomainService interface for testing.
// It provides a simple in-memory implementation of domain operations.
type MockErrDomainService struct{}

// ListDomains returns an empty list of domains for testing.
func (m *MockErrDomainService) ListDomains() ([]*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// GetDomain returns a mock domain entry for testing.
func (m *MockErrDomainService) GetDomain(_, _ string) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// CreateDomain creates a mock domain entry for testing.
func (m *MockErrDomainService) CreateDomain(_ *model.CreateDomainRequest) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// UpdateDomain updates a mock domain entry for testing.
func (m *MockErrDomainService) UpdateDomain(_ string, _ model.UpdateDomainRequest) (*model.DomainEntry, error) {
	return nil, fmt.Errorf("mock error")
}

// DeleteDomain simulates deleting a domain entry for testing.
func (m *MockErrDomainService) DeleteDomain(_ string, _ model.DeleteDomainRequest) error {
	return fmt.Errorf("mock error")
}

// Close performs cleanup for the mock service.
func (m *MockErrDomainService) Close() error {
	return nil
}
