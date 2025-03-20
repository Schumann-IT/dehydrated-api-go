package model

import (
	"regexp"
)

// IsValidDomain checks if a string is a valid domain name or wildcard domain
func IsValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Regular expression for domain validation
	// This pattern allows:
	// - Domain parts containing letters, numbers, and hyphens
	// - Hyphens cannot be at start or end of a part
	// - At least one dot (except for wildcard domains)
	// - Optional wildcard at the start of the first part
	// - TLD must be at least 2 characters
	pattern := `^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, domain)
	if err != nil {
		return false
	}

	return matched
}

// IsValidDomainEntry checks if a DomainEntry is valid by validating both the main domain
// and all alternative names
func IsValidDomainEntry(entry DomainEntry) bool {
	// Check main domain
	if !IsValidDomain(entry.Domain) {
		return false
	}

	// Check all alternative names
	for _, altName := range entry.AlternativeNames {
		if !IsValidDomain(altName) {
			return false
		}
	}

	return true
}
