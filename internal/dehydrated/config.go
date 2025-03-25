package dehydrated

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the dehydrated configuration
type Config struct {
	// User and group settings
	User  string `json:"user,omitempty" protobuf:"bytes,1,opt,name=user,proto3"`   // Which user should dehydrated run as
	Group string `json:"group,omitempty" protobuf:"bytes,2,opt,name=group,proto3"` // Which group should dehydrated run as

	// IP version settings
	IPVersion string // Resolve names to addresses of IP version only (4, 6)

	// Base directories
	BaseDir       string `json:"base_dir,omitempty" protobuf:"bytes,3,opt,name=base_dir,json=baseDir,proto3"`                   // Base directory for dehydrated
	CertDir       string `json:"cert_dir,omitempty" protobuf:"bytes,4,opt,name=cert_dir,json=certDir,proto3"`                   // Directory for certificates
	DomainsDir    string `json:"domains_dir,omitempty" protobuf:"bytes,5,opt,name=domains_dir,json=domainsDir,proto3"`          // Directory for domain configurations
	AccountsDir   string `json:"accounts_dir,omitempty" protobuf:"bytes,6,opt,name=accounts_dir,json=accountsDir,proto3"`       // Directory for account keys
	ChallengesDir string `json:"challenges_dir,omitempty" protobuf:"bytes,7,opt,name=challenges_dir,json=challengesDir,proto3"` // Directory for challenge files
	ChainCache    string `json:"chain_cache,omitempty" protobuf:"bytes,8,opt,name=chain_cache,json=chainCache,proto3"`          // Issuer chain cache directory

	// File paths
	DomainsFile string `json:"domains_file,omitempty" protobuf:"bytes,9,opt,name=domains_file,json=domainsFile,proto3"` // Path to domains.txt
	ConfigFile  string `json:"config_file,omitempty" protobuf:"bytes,10,opt,name=config_file,json=configFile,proto3"`   // Path to config file
	HookScript  string `json:"hook_script,omitempty" protobuf:"bytes,11,opt,name=hook_script,json=hookScript,proto3"`   // Path to hook script
	LockFile    string `json:"lock_file,omitempty" protobuf:"bytes,12,opt,name=lock_file,json=lockFile,proto3"`         // Path to lock file

	// OpenSSL settings
	OpensslConfig string `json:"openssl_config,omitempty" protobuf:"bytes,13,opt,name=openssl_config,json=opensslConfig,proto3"` // Path to openssl config file
	Openssl       string `json:"openssl,omitempty" protobuf:"bytes,14,opt,name=openssl,proto3"`                                  // Path to OpenSSL binary
	KeySize       int32  `json:"key_size,omitempty" protobuf:"varint,15,opt,name=key_size,json=keySize,proto3"`                  // Default keysize for private keys

	// ACME settings
	Ca             string `json:"ca,omitempty" protobuf:"bytes,16,opt,name=ca,proto3"`                                               // CA URL or preset
	OldCa          string `json:"old_ca,omitempty" protobuf:"bytes,17,opt,name=old_ca,json=oldCa,proto3"`                            // Path to old certificate authority
	AcceptTerms    bool   `json:"accept_terms,omitempty" protobuf:"varint,18,opt,name=accept_terms,json=acceptTerms,proto3"`         // Whether to accept CA terms
	Ipv4           bool   `json:"ipv4,omitempty" protobuf:"varint,19,opt,name=ipv4,proto3"`                                          // Resolve names to IPv4 only
	Ipv6           bool   `json:"ipv6,omitempty" protobuf:"varint,20,opt,name=ipv6,proto3"`                                          // Resolve names to IPv6 only
	PreferredChain string `json:"preferred_chain,omitempty" protobuf:"bytes,21,opt,name=preferred_chain,json=preferredChain,proto3"` // Alternative certificate chain
	Api            string `json:"api,omitempty" protobuf:"bytes,22,opt,name=api,proto3"`                                             // ACME API version

	// Certificate settings
	KeyAlgo            string `json:"key_algo,omitempty" protobuf:"bytes,23,opt,name=key_algo,json=keyAlgo,proto3"`                                     // Public key algorithm (rsa, prime256v1, secp384r1)
	RenewDays          int32  `json:"renew_days,omitempty" protobuf:"varint,24,opt,name=renew_days,json=renewDays,proto3"`                              // Days before renewal
	ForceRenew         bool   `json:"force_renew,omitempty" protobuf:"varint,25,opt,name=force_renew,json=forceRenew,proto3"`                           // Force certificate renewal
	ForceValidation    bool   `json:"force_validation,omitempty" protobuf:"varint,26,opt,name=force_validation,json=forceValidation,proto3"`            // Force domain validation
	PrivateKeyRenew    bool   `json:"private_key_renew,omitempty" protobuf:"varint,27,opt,name=private_key_renew,json=privateKeyRenew,proto3"`          // Regenerate private keys on renewal
	PrivateKeyRollover bool   `json:"private_key_rollover,omitempty" protobuf:"varint,28,opt,name=private_key_rollover,json=privateKeyRollover,proto3"` // Create extra private key for rollover

	// Challenge settings
	ChallengeType string `json:"challenge_type,omitempty" protobuf:"bytes,29,opt,name=challenge_type,json=challengeType,proto3"` // Challenge type (http-01, dns-01, tls-alpn-01)
	WellKnownDir  string `json:"well_known_dir,omitempty" protobuf:"bytes,30,opt,name=well_known_dir,json=wellKnownDir,proto3"`  // Directory for http-01 challenge
	AlpnDir       string `json:"alpn_dir,omitempty" protobuf:"bytes,31,opt,name=alpn_dir,json=alpnDir,proto3"`                   // Directory for tls-alpn-01 challenge
	HookChain     bool   `json:"hook_chain,omitempty" protobuf:"varint,32,opt,name=hook_chain,json=hookChain,proto3"`            // Chain challenge arguments together

	// OCSP settings
	OcspMustStaple bool  `json:"ocsp_must_staple,omitempty" protobuf:"varint,33,opt,name=ocsp_must_staple,json=ocspMustStaple,proto3"` // Add CSR-flag indicating OCSP stapling mandatory
	OcspFetch      bool  `json:"ocsp_fetch,omitempty" protobuf:"varint,34,opt,name=ocsp_fetch,json=ocspFetch,proto3"`                  // Fetch OCSP responses
	OcspDays       int32 `json:"ocsp_days,omitempty" protobuf:"varint,35,opt,name=ocsp_days,json=ocspDays,proto3"`                     // OCSP refresh interval

	// Other settings
	NoLock       bool   `json:"no_lock,omitempty" protobuf:"varint,36,opt,name=no_lock,json=noLock,proto3"`                  // Don't use lockfile
	KeepGoing    bool   `json:"keep_going,omitempty" protobuf:"varint,37,opt,name=keep_going,json=keepGoing,proto3"`         // Continue after errors
	FullChain    bool   `json:"full_chain,omitempty" protobuf:"varint,38,opt,name=full_chain,json=fullChain,proto3"`         // Print full chain
	Ocsp         bool   `json:"ocsp,omitempty" protobuf:"varint,39,opt,name=ocsp,proto3"`                                    // Enable OCSP stapling
	AutoCleanup  bool   `json:"auto_cleanup,omitempty" protobuf:"varint,40,opt,name=auto_cleanup,json=autoCleanup,proto3"`   // Automatic cleanup
	ContactEmail string `json:"contact_email,omitempty" protobuf:"bytes,41,opt,name=contact_email,json=contactEmail,proto3"` // E-mail to use during registration
	CurlOpts     string `json:"curl_opts,omitempty" protobuf:"bytes,42,opt,name=curl_opts,json=curlOpts,proto3"`             // Extra options passed to curl
	ConfigD      string `json:"config_d,omitempty" protobuf:"bytes,43,opt,name=config_d,json=configD,proto3"`                // Directory containing additional config files
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
		Ca:              "letsencrypt",
		OldCa:           "https://acme-v01.api.letsencrypt.org/directory",
		RenewDays:       30,
		KeySize:         4096,
		KeyAlgo:         "rsa",
		ChallengeType:   "http-01",
		WellKnownDir:    "/var/www/dehydrated",
		LockFile:        "dehydrated.lock",
		Openssl:         "openssl",
		PrivateKeyRenew: true,
		HookChain:       false,
		OcspDays:        5,
		ChainCache:      "chains",
		Api:             "auto",
	}
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Group:    "www-data",
		Ca:       "https://acme-v02.api.letsencrypt.org/directory",
		Openssl:  "openssl",
		OcspDays: 5,
		Api:      "v2",
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
			c.Ca = value
		case "OLDCA":
			c.OldCa = value
		case "ACCEPT_TERMS":
			c.AcceptTerms = value == "yes"
		case "IPV4":
			c.Ipv4 = value == "yes"
		case "IPV6":
			c.Ipv6 = value == "yes"
		case "PREFERRED_CHAINS":
			c.PreferredChain = value
		case "API":
			c.Api = value
		case "KEY_ALGO":
			c.KeyAlgo = value
		case "KEY_SIZE":
			if size, err := strconv.Atoi(value); err == nil {
				c.KeySize = int32(size)
			}
		case "RENEW_DAYS":
			if days, err := strconv.Atoi(value); err == nil {
				c.RenewDays = int32(days)
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
			c.AlpnDir = value
		case "LOCKFILE":
			c.LockFile = value
		case "NO_LOCK":
			c.NoLock = value == "yes"
		case "KEEP_GOING":
			c.KeepGoing = value == "yes"
		case "FULL_CHAIN":
			c.FullChain = value == "yes"
		case "OCSP":
			c.Ocsp = value == "yes"
		case "OCSP_MUST_STAPLE":
			c.OcspMustStaple = value == "yes"
		case "OCSP_FETCH":
			c.OcspFetch = value == "yes"
		case "OCSP_DAYS":
			if days, err := strconv.Atoi(value); err == nil {
				c.OcspDays = int32(days)
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
			c.OpensslConfig = value
		case "OPENSSL":
			c.Openssl = value
		case "GROUP":
			c.Group = value
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
	if c.OpensslConfig != "" {
		c.OpensslConfig = c.ensureAbs(c.OpensslConfig)
	}
}
