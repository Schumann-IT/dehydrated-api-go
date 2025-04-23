package model

import (
	"regexp"
)

// IsValidDomain checks if a string is a valid domain name or wildcard domain.
// It validates the domain against a regular expression that enforces the following rules:
// - Domain parts can contain letters, numbers, and hyphens
// - Hyphens cannot be at the start or end of a part
// - At least one dot is required (except for wildcard domains)
// - Optional wildcard at the start of the first part
// - TLD must be at least 2 characters
// Returns true if the domain is valid, false otherwise.
func IsValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Regular expression for domain validation
	pattern := `^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, domain)
	if err != nil {
		return false
	}

	return matched
}

// IsValidDomainEntry checks if a DomainEntry is valid by validating its domain field.
// It ensures that the domain name follows the standard domain naming conventions.
// Returns true if the domain entry is valid, false otherwise.
func IsValidDomainEntry(entry DomainEntry) bool {
	return IsValidDomain(entry.Domain)
}
