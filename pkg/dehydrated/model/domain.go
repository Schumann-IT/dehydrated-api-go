package model

// DomainEntry represents a domain entry in the domains.txt file
type DomainEntry struct {
	Domain           string                 `json:"domain"`
	AlternativeNames []string               `json:"alternative_names,omitempty"`
	Alias            string                 `json:"alias,omitempty"`
	Enabled          bool                   `json:"enabled"`
	Comment          string                 `json:"comment,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDomainRequest represents the request body for creating a domain
type CreateDomainRequest struct {
	Domain           string   `json:"domain" validate:"required"`
	AlternativeNames []string `json:"alternative_names,omitempty"`
	Alias            string   `json:"alias,omitempty"`
	Enabled          bool     `json:"enabled"`
	Comment          string   `json:"comment,omitempty"`
}

// UpdateDomainRequest represents the request body for updating a domain
type UpdateDomainRequest struct {
	AlternativeNames []string `json:"alternative_names,omitempty"`
	Alias            string   `json:"alias,omitempty"`
	Enabled          bool     `json:"enabled"`
	Comment          string   `json:"comment,omitempty"`
}

// DomainResponse represents the response for domain operations
type DomainResponse struct {
	Success bool        `json:"success"`
	Data    DomainEntry `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// DomainsResponse represents the response for listing domains
type DomainsResponse struct {
	Success bool          `json:"success"`
	Data    []DomainEntry `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}
