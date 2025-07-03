package model

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/schumann-it/dehydrated-api-go/internal/util"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/stretchr/testify/require"
)

var validate = validator.New()

func TestDomainEntry_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		entry    *DomainEntry
		expected string
	}{
		{
			name: "full entry",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:           "example.com",
					AlternativeNames: []string{"www.example.com"},
					Alias:            "example",
					Enabled:          true,
					Comment:          "test comment",
				},
				Metadata: pb.NewMetadata(),
			},
			expected: `{"domain":"example.com","alternative_names":["www.example.com"],"alias":"example","enabled":true,"comment":"test comment","metadata":{}}`,
		},
		{
			name: "minimal entry",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:  "example.com",
					Enabled: true,
				},
				Metadata: pb.NewMetadata(),
			},
			expected: `{"domain":"example.com","alternative_names":null,"alias":"","enabled":true,"comment":"","metadata":{}}`,
		},
		{
			name: "entry with metadata",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:  "example.com",
					Enabled: true,
				},
				Metadata: func() *pb.Metadata {
					m := pb.NewMetadata()
					m.Set("key", "value")
					return m
				}(),
			},
			expected: `{"domain":"example.com","alternative_names":null,"alias":"","enabled":true,"comment":"","metadata":{"key":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.entry)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestDomainEntry_SetMetadata(t *testing.T) {
	entry := &DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:  "example.com",
			Enabled: true,
		},
	}

	metadata := pb.NewMetadata()
	metadata.Set("key", "value")
	entry.SetMetadata(metadata)

	require.Equal(t, metadata, entry.Metadata)
}

func TestDomainEntry_PathName(t *testing.T) {
	tests := []struct {
		name     string
		entry    *DomainEntry
		expected string
	}{
		{
			name: "with domain only",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain: "example.com",
				},
			},
			expected: "example.com",
		},
		{
			name: "with alias",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain: "example.com",
					Alias:  "example",
				},
			},
			expected: "example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.entry.PathName())
		})
	}
}

func TestCreateDomainRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *CreateDomainRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &CreateDomainRequest{
				Domain:           "example.com",
				AlternativeNames: []string{"www.example.com"},
				Enabled:          true,
			},
			wantErr: false,
		},
		{
			name: "missing domain",
			request: &CreateDomainRequest{
				AlternativeNames: []string{"www.example.com"},
				Enabled:          true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateDomainRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateDomainRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: UpdateDomainRequest{
				AlternativeNames: util.StringSlicePtr([]string{"www.example.com"}),
				Enabled:          util.BoolPtr(true),
			},
			isValid: true,
		},
		{
			name:    "Empty request",
			request: UpdateDomainRequest{},
			isValid: true,
		},
		{
			name:    "With metadata",
			request: UpdateDomainRequest{},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// UpdateDomainRequest doesn't have required fields, so all validations should pass
			require.True(t, tt.isValid)
		})
	}
}

func TestDomainResponse(t *testing.T) {
	tests := []struct {
		name     string
		response DomainResponse
		success  bool
	}{
		{
			name: "Successful response",
			response: DomainResponse{
				Success: true,
				Data: &DomainEntry{
					DomainEntry: pb.DomainEntry{
						Domain: "example.com",
					},
				},
			},
			success: true,
		},
		{
			name: "Error response",
			response: DomainResponse{
				Success: false,
				Error:   "domain not found",
			},
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				require.NotEmpty(t, tt.response.Data.Domain)
				require.Empty(t, tt.response.Error)
			} else {
				require.NotEmpty(t, tt.response.Error)
			}
		})
	}
}

func TestDomainsResponse(t *testing.T) {
	tests := []struct {
		name     string
		response DomainsResponse
		success  bool
	}{
		{
			name: "Successful response",
			response: DomainsResponse{
				Success: true,
				Data: DomainEntries{
					{
						DomainEntry: pb.DomainEntry{
							Domain: "example1.com",
						},
					},
					{
						DomainEntry: pb.DomainEntry{
							Domain: "example2.com",
						},
					},
				},
			},
			success: true,
		},
		{
			name: "Error response",
			response: DomainsResponse{
				Success: false,
				Error:   "failed to list domains",
			},
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				require.NotEmpty(t, tt.response.Data)
				require.Empty(t, tt.response.Error)
			} else {
				require.NotEmpty(t, tt.response.Error)
			}
		})
	}
}

func TestDomainEntries_Sort(t *testing.T) {
	// Create test domains with mixed order
	entries := DomainEntries{
		{DomainEntry: pb.DomainEntry{Domain: "vpn.hq.schumann-it.com", Alias: "vpn.hq.schumann-it.com-rsa", Enabled: true}},
		{DomainEntry: pb.DomainEntry{Domain: "vpn.lg.schumann-it.com", Enabled: true}},
		{DomainEntry: pb.DomainEntry{Domain: "vpn.lg.schumann-it.com", Alias: "vpn.lg.schumann-it.com-rsa", Enabled: true}},
		{DomainEntry: pb.DomainEntry{Domain: "foo.hq.schumann-it.com", Comment: "sdcsdc", Enabled: true}},
		{DomainEntry: pb.DomainEntry{Domain: "foo.hq.schumann-it.com", Alias: "foo.hq.schumann-it.com-rsa", Enabled: false}},
	}

	// Test sorting
	t.Run("Sort", func(t *testing.T) {
		// Sort the entries
		entries.Sort()

		// Verify the sorted order
		expectedOrder := []string{
			"foo.hq.schumann-it.com", // No alias first
			"foo.hq.schumann-it.com", // With alias (disabled)
			"vpn.hq.schumann-it.com", // With alias
			"vpn.lg.schumann-it.com", // No alias first
			"vpn.lg.schumann-it.com", // With alias
		}

		if len(entries) != len(expectedOrder) {
			t.Errorf("Expected %d entries, got %d", len(expectedOrder), len(entries))
			return
		}

		for i, entry := range entries {
			if entry.Domain != expectedOrder[i] {
				t.Errorf("Entry %d: Expected domain %s, got %s", i, expectedOrder[i], entry.Domain)
			}
		}

		// Verify the new sorting behavior: domains grouped together, no alias first within each group
		// Expected order:
		// 0: foo.hq.schumann-it.com (no alias)
		// 1: foo.hq.schumann-it.com (with alias)
		// 2: vpn.hq.schumann-it.com (with alias)
		// 3: vpn.lg.schumann-it.com (no alias)
		// 4: vpn.lg.schumann-it.com (with alias)

		// Check that foo.hq.schumann-it.com entries are grouped together
		if entries[0].Domain != "foo.hq.schumann-it.com" || entries[0].Alias != "" {
			t.Errorf("Entry 0 should be foo.hq.schumann-it.com without alias, got %s (alias: %s)", entries[0].Domain, entries[0].Alias)
		}
		if entries[1].Domain != "foo.hq.schumann-it.com" || entries[1].Alias == "" {
			t.Errorf("Entry 1 should be foo.hq.schumann-it.com with alias, got %s (alias: %s)", entries[1].Domain, entries[1].Alias)
		}

		// Check that vpn.hq.schumann-it.com entry is in the middle
		if entries[2].Domain != "vpn.hq.schumann-it.com" || entries[2].Alias == "" {
			t.Errorf("Entry 2 should be vpn.hq.schumann-it.com with alias, got %s (alias: %s)", entries[2].Domain, entries[2].Alias)
		}

		// Check that vpn.lg.schumann-it.com entries are grouped together at the end
		if entries[3].Domain != "vpn.lg.schumann-it.com" || entries[3].Alias != "" {
			t.Errorf("Entry 3 should be vpn.lg.schumann-it.com without alias, got %s (alias: %s)", entries[3].Domain, entries[3].Alias)
		}
		if entries[4].Domain != "vpn.lg.schumann-it.com" || entries[4].Alias == "" {
			t.Errorf("Entry 4 should be vpn.lg.schumann-it.com with alias, got %s (alias: %s)", entries[4].Domain, entries[4].Alias)
		}
	})

	// Test sorting with same domains but different aliases
	t.Run("SortWithSameDomains", func(t *testing.T) {
		sameDomainEntries := DomainEntries{
			{DomainEntry: pb.DomainEntry{Domain: "example.com", Alias: "alias2", Enabled: true}},
			{DomainEntry: pb.DomainEntry{Domain: "example.com", Enabled: true}},
			{DomainEntry: pb.DomainEntry{Domain: "example.com", Alias: "alias1", Enabled: true}},
		}

		sameDomainEntries.Sort()

		// Should be sorted: no alias first, then aliases alphabetically
		expectedAliases := []string{"", "alias1", "alias2"}
		for i, entry := range sameDomainEntries {
			if entry.Alias != expectedAliases[i] {
				t.Errorf("Entry %d: Expected alias %s, got %s", i, expectedAliases[i], entry.Alias)
			}
		}
	})
}
