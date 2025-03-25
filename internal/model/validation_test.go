package model

import "testing"

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{"Valid domain", "example.com", true},
		{"Valid domain with subdomain", "www.example.com", true},
		{"Valid domain with multiple subdomains", "mail.www.example.com", true},
		{"Valid wildcard domain", "*.example.com", true},
		{"Valid wildcard domain with subdomain", "*.www.example.com", true},
		{"Valid domain with hyphens", "my-domain.com", true},
		{"Valid domain with numbers", "example123.com", true},
		{"Empty domain", "", false},
		{"Domain without dots", "example", false},
		{"Domain starting with dot", ".example.com", false},
		{"Domain ending with dot", "example.com.", false},
		{"Domain with invalid characters", "example@domain.com", false},
		{"Domain part starting with hyphen", "-example.com", false},
		{"Domain part ending with hyphen", "example-.com", false},
		{"Multiple wildcards", "*.example.*.com", false},
		{"Wildcard in middle", "example.*.com", false},
		{"Wildcard at end", "example.com.*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("IsValidDomain(%q) = %v; want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestIsValidDomainEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    DomainEntry
		expected bool
	}{
		{
			name: "Valid entry with valid domain",
			entry: DomainEntry{
				Domain: "example.com",
			},
			expected: true,
		},
		{
			name: "Valid entry with wildcard domain",
			entry: DomainEntry{
				Domain: "*.example.com",
			},
			expected: true,
		},
		{
			name: "Invalid entry with invalid domain",
			entry: DomainEntry{
				Domain: "invalid@domain.com",
			},
			expected: false,
		},
		{
			name: "Invalid entry with empty domain",
			entry: DomainEntry{
				Domain: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDomainEntry(tt.entry)
			if result != tt.expected {
				t.Errorf("IsValidDomainEntry(%v) = %v; want %v", tt.entry, result, tt.expected)
			}
		})
	}
}
