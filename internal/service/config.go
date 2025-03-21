package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the dehydrated configuration
type Config struct {
	// User and group settings
	User  string // Which user should dehydrated run as
	Group string // Which group should dehydrated run as

	// IP version settings
	IPVersion string // Resolve names to addresses of IP version only (4, 6)

	// Base directories
	BaseDir       string // Base directory for dehydrated
	CertDir       string // Directory for certificates
	DomainsDir    string // Directory for domain configurations
	AccountsDir   string // Directory for account keys
	ChallengesDir string // Directory for challenge files
	ChainCache    string // Issuer chain cache directory

	// File paths
	DomainsFile string // Path to domains.txt
	ConfigFile  string // Path to config file
	HookScript  string // Path to hook script
	LockFile    string // Path to lock file

	// OpenSSL settings
	OpenSSLConfig string // Path to openssl config file
	OpenSSL       string // Path to OpenSSL binary
	KeySize       int    // Default keysize for private keys

	// ACME settings
	CA             string // CA URL or preset
	OldCA          string // Path to old certificate authority
	AcceptTerms    bool   // Whether to accept CA terms
	IPV4           bool   // Resolve names to IPv4 only
	IPV6           bool   // Resolve names to IPv6 only
	PreferredChain string // Alternative certificate chain
	API            string // ACME API version

	// Certificate settings
	KeyAlgo            string // Public key algorithm (rsa, prime256v1, secp384r1)
	RenewDays          int    // Days before renewal
	ForceRenew         bool   // Force certificate renewal
	ForceValidation    bool   // Force domain validation
	PrivateKeyRenew    bool   // Regenerate private keys on renewal
	PrivateKeyRollover bool   // Create extra private key for rollover

	// Challenge settings
	ChallengeType string // Challenge type (http-01, dns-01, tls-alpn-01)
	WellKnownDir  string // Directory for http-01 challenge
	ALPNDir       string // Directory for tls-alpn-01 challenge
	HookChain     bool   // Chain challenge arguments together

	// OCSP settings
	OCSPMustStaple bool // Add CSR-flag indicating OCSP stapling mandatory
	OCSPFetch      bool // Fetch OCSP responses
	OCSPDays       int  // OCSP refresh interval

	// Other settings
	NoLock       bool   // Don't use lockfile
	KeepGoing    bool   // Continue after errors
	FullChain    bool   // Print full chain
	OCSP         bool   // Enable OCSP stapling
	AutoCleanup  bool   // Automatic cleanup
	ContactEmail string // E-mail to use during registration
	CurlOpts     string // Extra options passed to curl
	ConfigD      string // Directory containing additional config files
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		BaseDir:         ".",
		CertDir:         "certs",
		DomainsDir:      "domains",
		AccountsDir:     "accounts",
		ChallengesDir:   "acme-challenges",
		DomainsFile:     "domains.txt",
		CA:              "letsencrypt",
		OldCA:           "https://acme-v01.api.letsencrypt.org/directory",
		RenewDays:       30,
		KeySize:         4096,
		KeyAlgo:         "rsa",
		ChallengeType:   "http-01",
		WellKnownDir:    "/var/www/dehydrated",
		LockFile:        "dehydrated.lock",
		OpenSSL:         "openssl",
		PrivateKeyRenew: true,
		HookChain:       false,
		OCSPDays:        5,
		ChainCache:      "chains",
		API:             "auto",
	}
}

// WithBaseDir sets the base directory for the config
func (c *Config) WithBaseDir(baseDir string) *Config {
	c.BaseDir = baseDir
	return c
}

// WithConfigFile sets the path to the config file
func (c *Config) WithConfigFile(configFile string) *Config {
	c.ConfigFile = configFile
	return c
}

// Load loads the configuration from files
func (c *Config) Load() *Config {
	// make baseDir absolute
	if !filepath.IsAbs(c.BaseDir) {
		abs, _ := filepath.Abs(c.BaseDir)
		c.BaseDir = abs
	}

	c.load()
	c.resolvePaths()

	return c
}

func (c *Config) findAndSetConfigFile() {
	if c.ConfigFile == "" {
		files := []string{"config", "config.sh"}
		for _, file := range files {
			if _, err := os.Stat(filepath.Join(c.BaseDir, file)); err == nil {
				c.ConfigFile = filepath.Join(c.BaseDir, file)
				return
			}
		}
	}
}

// load loads configuration from a config file if it exists
func (c *Config) load() {
	if c.ConfigFile == "" {
		c.findAndSetConfigFile()
	}

	if c.ConfigFile != "" {
		c.ConfigFile = c.ensureAbs(c.ConfigFile)
	}

	if c.ConfigFile == "" {
		return
	}

	// Read config file
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return
	}

	// Parse config file
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for export statements
		line = strings.TrimPrefix(line, "export ")

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
			c.BaseDir = value
		case "CERTDIR":
			c.CertDir = value
		case "DOMAINSD":
			c.DomainsDir = value
		case "ACCOUNTDIR":
			c.AccountsDir = value
		case "CHALLENGEDIR":
			c.ChallengesDir = value
		case "DOMAINS_TXT":
			c.DomainsFile = value
		case "HOOK":
			c.HookScript = value
		case "CA":
			c.CA = value
		case "OLDCA":
			c.OldCA = value
		case "ACCEPT_TERMS":
			c.AcceptTerms = value == "yes"
		case "IPV4":
			c.IPV4 = value == "yes"
		case "IPV6":
			c.IPV6 = value == "yes"
		case "PREFERRED_CHAINS":
			c.PreferredChain = value
		case "API":
			c.API = value
		case "KEY_ALGO":
			c.KeyAlgo = value
		case "KEY_SIZE":
			if size, err := strconv.Atoi(value); err == nil {
				c.KeySize = size
			}
		case "RENEW_DAYS":
			if days, err := strconv.Atoi(value); err == nil {
				c.RenewDays = days
			}
		case "FORCE_RENEW":
			c.ForceRenew = value == "yes"
		case "FORCE_VALIDATION":
			c.ForceValidation = value == "yes"
		case "CHALLENGETYPE":
			c.ChallengeType = value
		case "WELLKNOWN":
			c.WellKnownDir = value
		case "ALPNCERTDIR":
			c.ALPNDir = value
		case "LOCKFILE":
			c.LockFile = value
		case "NO_LOCK":
			c.NoLock = value == "yes"
		case "KEEP_GOING":
			c.KeepGoing = value == "yes"
		case "FULL_CHAIN":
			c.FullChain = value == "yes"
		case "OCSP":
			c.OCSP = value == "yes"
		case "OCSP_MUST_STAPLE":
			c.OCSPMustStaple = value == "yes"
		case "OCSP_FETCH":
			c.OCSPFetch = value == "yes"
		case "OCSP_DAYS":
			if days, err := strconv.Atoi(value); err == nil {
				c.OCSPDays = days
			}
		case "AUTO_CLEANUP":
			c.AutoCleanup = value == "yes"
		case "CONTACT_EMAIL":
			c.ContactEmail = value
		case "CURL_OPTS":
			c.CurlOpts = value
		case "CONFIG_D":
			c.ConfigD = value
		case "OPENSSL_CONFIG":
			c.OpenSSLConfig = value
		case "OPENSSL":
			c.OpenSSL = value
		case "PRIVATE_KEY_RENEW":
			c.PrivateKeyRenew = value == "yes"
		case "PRIVATE_KEY_ROLLOVER":
			c.PrivateKeyRollover = value == "yes"
		case "HOOK_CHAIN":
			c.HookChain = value == "yes"
		case "CHAIN_CACHE":
			c.ChainCache = value
		}
	}

	// Resolve relative paths
	c.resolvePaths()
}

func (c *Config) ensureAbs(p string) string {
	if !filepath.IsAbs(p) {
		return filepath.Join(c.BaseDir, p)
	}

	return p
}

// resolvePaths converts relative paths to absolute paths
func (c *Config) resolvePaths() {
	c.BaseDir = c.ensureAbs(c.BaseDir)
	c.CertDir = c.ensureAbs(c.CertDir)
	c.DomainsDir = c.ensureAbs(c.DomainsDir)
	c.AccountsDir = c.ensureAbs(c.AccountsDir)
	c.ChallengesDir = c.ensureAbs(c.ChallengesDir)
	c.DomainsFile = c.ensureAbs(c.DomainsFile)

	if c.HookScript != "" {
		c.HookScript = c.ensureAbs(c.HookScript)
	}
	if c.LockFile != "" {
		c.LockFile = c.ensureAbs(c.LockFile)
	}
	if c.OpenSSLConfig != "" {
		c.OpenSSLConfig = c.ensureAbs(c.OpenSSLConfig)
	}
}
