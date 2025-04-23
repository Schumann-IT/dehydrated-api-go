package auth

// Config holds the Azure AD authentication configuration
type Config struct {
	// TenantID is the Azure AD tenant ID
	TenantID string `yaml:"tenantId"`

	// ClientID is the Azure AD application (client) ID
	ClientID string `yaml:"clientId"`

	// Authority is the Azure AD authority URL (e.g., https://login.microsoftonline.com/{tenantId})
	Authority string `yaml:"authority"`

	// AllowedAudiences is a list of allowed audience values in the token
	AllowedAudiences []string `yaml:"allowedAudiences"`

	// EnableManagedIdentity enables managed identity authentication
	EnableManagedIdentity bool `yaml:"enableManagedIdentity"`

	// EnableServicePrincipal enables service principal authentication
	EnableServicePrincipal bool `yaml:"enableServicePrincipal"`

	// EnableUserAuthentication enables user authentication
	EnableUserAuthentication bool `yaml:"enableUserAuthentication"`
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		EnableManagedIdentity:    true,
		EnableServicePrincipal:   true,
		EnableUserAuthentication: true,
	}
}
