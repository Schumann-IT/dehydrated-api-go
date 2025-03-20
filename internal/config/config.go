package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the dehydrated configuration
type Config struct {
	// Base directories
	BaseDir       string // Base directory for dehydrated
	CertDir       string // Directory for certificates
	DomainsDir    string // Directory for domain configurations
	AccountsDir   string // Directory for account keys
	ChallengesDir string // Directory for challenge files

	// File paths
	DomainsFile string // Path to domains.txt
	ConfigFile  string // Path to config file
	HookScript  string // Path to hook script

	// ACME settings
	CA             string // CA URL or preset
	AcceptTerms    bool   // Whether to accept CA terms
	IPV4           bool   // Resolve names to IPv4 only
	IPV6           bool   // Resolve names to IPv6 only
	PreferredChain string // Alternative certificate chain

	// Certificate settings
	KeyAlgo         string // Public key algorithm (rsa, prime256v1, secp384r1)
	RenewDays       int    // Days before renewal
	ForceRenew      bool   // Force certificate renewal
	ForceValidation bool   // Force domain validation

	// Challenge settings
	ChallengeType string // Challenge type (http-01, dns-01, tls-alpn-01)
	WellKnownDir  string // Directory for http-01 challenge
	ALPNDir       string // Directory for tls-alpn-01 challenge

	// Other settings
	LockFile  string // Path to lock file
	NoLock    bool   // Don't use lockfile
	KeepGoing bool   // Continue after errors
	FullChain bool   // Print full chain
	OCSP      bool   // Enable OCSP stapling
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		BaseDir:       "/etc/dehydrated",
		CertDir:       "certs",
		DomainsDir:    "domains",
		AccountsDir:   "accounts",
		ChallengesDir: "acme-challenges",
		DomainsFile:   "domains.txt",
		RenewDays:     30,
		KeyAlgo:       "rsa",
		ChallengeType: "http-01",
		WellKnownDir:  "/var/www/dehydrated",
		LockFile:      "dehydrated.lock",
	}
}

// FindConfigFile searches for the config file in the standard locations
func FindConfigFile() (string, error) {
	locations := []string{
		"/etc/dehydrated/config",
		"/usr/local/etc/dehydrated/config",
		"config",
		filepath.Join(os.Getenv("PWD"), "config"),
	}

	for _, loc := range locations {
		// Convert relative paths to absolute
		absPath := loc
		if !filepath.IsAbs(loc) {
			if abs, err := filepath.Abs(loc); err == nil {
				absPath = abs
			}
		}

		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", os.ErrNotExist
}

// LoadConfig reads and parses the dehydrated config file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()
	config.ConfigFile = configPath

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config file
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		switch key {
		case "BASEDIR":
			config.BaseDir = value
		case "CERTDIR":
			config.CertDir = value
		case "DOMAINSD":
			config.DomainsDir = value
		case "ACCOUNTDIR":
			config.AccountsDir = value
		case "CHALLENGEDIR":
			config.ChallengesDir = value
		case "DOMAINS_TXT":
			config.DomainsFile = value
		case "HOOK":
			config.HookScript = value
		case "CA":
			config.CA = value
		case "ACCEPT_TERMS":
			config.AcceptTerms = value == "yes"
		case "IPV4":
			config.IPV4 = value == "yes"
		case "IPV6":
			config.IPV6 = value == "yes"
		case "PREFERRED_CHAINS":
			config.PreferredChain = value
		case "KEY_ALGO":
			config.KeyAlgo = value
		case "RENEW_DAYS":
			if days, err := strconv.Atoi(value); err == nil {
				config.RenewDays = days
			}
		case "FORCE_RENEW":
			config.ForceRenew = value == "yes"
		case "FORCE_VALIDATION":
			config.ForceValidation = value == "yes"
		case "CHALLENGETYPE":
			config.ChallengeType = value
		case "WELLKNOWN":
			config.WellKnownDir = value
		case "ALPNCERTDIR":
			config.ALPNDir = value
		case "LOCKFILE":
			config.LockFile = value
		case "NO_LOCK":
			config.NoLock = value == "yes"
		case "KEEP_GOING":
			config.KeepGoing = value == "yes"
		case "FULL_CHAIN":
			config.FullChain = value == "yes"
		case "OCSP":
			config.OCSP = value == "yes"
		}
	}

	// Resolve relative paths
	config.resolvePaths()

	return config, nil
}

// resolvePaths converts relative paths to absolute paths
func (c *Config) resolvePaths() {
	baseDir := c.BaseDir
	if !filepath.IsAbs(baseDir) {
		if abs, err := filepath.Abs(baseDir); err == nil {
			baseDir = abs
		}
	}

	// Resolve all paths relative to baseDir
	c.CertDir = filepath.Join(baseDir, c.CertDir)
	c.DomainsDir = filepath.Join(baseDir, c.DomainsDir)
	c.AccountsDir = filepath.Join(baseDir, c.AccountsDir)
	c.ChallengesDir = filepath.Join(baseDir, c.ChallengesDir)
	c.DomainsFile = filepath.Join(baseDir, c.DomainsFile)
	if c.HookScript != "" && !filepath.IsAbs(c.HookScript) {
		c.HookScript = filepath.Join(baseDir, c.HookScript)
	}
	if c.LockFile != "" && !filepath.IsAbs(c.LockFile) {
		c.LockFile = filepath.Join(baseDir, c.LockFile)
	}
}
