package auth

// Config holds the configuration for Azure AD authentication middleware
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

	// EnableSignatureValidation enables JWT signature validation
	// When enabled, the middleware will fetch and validate Azure AD public keys
	EnableSignatureValidation bool `yaml:"enableSignatureValidation"`

	// KeyCacheTTL is the time-to-live for the public key cache (e.g., "24h", "1h")
	// Defaults to 24 hours if not specified
	KeyCacheTTL string `yaml:"keyCacheTTL"`
}

// NewConfig creates a new Config instance with default values
// This function initializes the configuration with reasonable defaults
// for Azure AD authentication, including enabling managed identity,
// service principal, user authentication, and signature validation.
func NewConfig() *Config {
	return &Config{
		EnableManagedIdentity:     true,
		EnableServicePrincipal:    true,
		EnableUserAuthentication:  true,
		EnableSignatureValidation: true,  // Enable signature validation by default
		KeyCacheTTL:               "24h", // Default to 24 hours
	}
}
