// Package dehydrated provides functionality for working with the dehydrated ACME client.
// It includes configuration management, path resolution, and integration with the dehydrated script.
package dehydrated

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
)

// Config represents the dehydrated configuration
type Config struct {
	pb.DehydratedConfig
}

// MarshalJSON implements the json.Marshaler interface to ensure all fields are included
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config // Create an alias to avoid recursion
	return json.Marshal(&struct {
		User               string `json:"user"`
		Group              string `json:"group"`
		BaseDir            string `json:"base_dir"`
		CertDir            string `json:"cert_dir"`
		DomainsDir         string `json:"domains_dir"`
		AccountsDir        string `json:"accounts_dir"`
		ChallengesDir      string `json:"challenges_dir"`
		ChainCache         string `json:"chain_cache"`
		DomainsFile        string `json:"domains_file"`
		ConfigFile         string `json:"config_file"`
		HookScript         string `json:"hook_script"`
		LockFile           string `json:"lock_file"`
		OpensslConfig      string `json:"openssl_config"`
		Openssl            string `json:"openssl"`
		KeySize            int32  `json:"key_size"`
		Ca                 string `json:"ca"`
		OldCa              string `json:"old_ca"`
		AcceptTerms        bool   `json:"accept_terms"`
		Ipv4               bool   `json:"ipv4"`
		Ipv6               bool   `json:"ipv6"`
		PreferredChain     string `json:"preferred_chain"`
		Api                string `json:"api"`
		KeyAlgo            string `json:"key_algo"`
		RenewDays          int32  `json:"renew_days"`
		ForceRenew         bool   `json:"force_renew"`
		ForceValidation    bool   `json:"force_validation"`
		PrivateKeyRenew    bool   `json:"private_key_renew"`
		PrivateKeyRollover bool   `json:"private_key_rollover"`
		ChallengeType      string `json:"challenge_type"`
		WellKnownDir       string `json:"well_known_dir"`
		AlpnDir            string `json:"alpn_dir"`
		HookChain          bool   `json:"hook_chain"`
		OcspMustStaple     bool   `json:"ocsp_must_staple"`
		OcspFetch          bool   `json:"ocsp_fetch"`
		OcspDays           int32  `json:"ocsp_days"`
		NoLock             bool   `json:"no_lock"`
		KeepGoing          bool   `json:"keep_going"`
		FullChain          bool   `json:"full_chain"`
		Ocsp               bool   `json:"ocsp"`
		AutoCleanup        bool   `json:"auto_cleanup"`
		ContactEmail       string `json:"contact_email"`
		CurlOpts           string `json:"curl_opts"`
		ConfigD            string `json:"config_d"`
	}{
		User:               c.GetUser(),
		Group:              c.GetGroup(),
		BaseDir:            c.GetBaseDir(),
		CertDir:            c.GetCertDir(),
		DomainsDir:         c.GetDomainsDir(),
		AccountsDir:        c.GetAccountsDir(),
		ChallengesDir:      c.GetChallengesDir(),
		ChainCache:         c.GetChainCache(),
		DomainsFile:        c.GetDomainsFile(),
		ConfigFile:         c.GetConfigFile(),
		HookScript:         c.GetHookScript(),
		LockFile:           c.GetLockFile(),
		OpensslConfig:      c.GetOpensslConfig(),
		Openssl:            c.GetOpenssl(),
		KeySize:            c.GetKeySize(),
		Ca:                 c.GetCa(),
		OldCa:              c.GetOldCa(),
		AcceptTerms:        c.GetAcceptTerms(),
		Ipv4:               c.GetIpv4(),
		Ipv6:               c.GetIpv6(),
		PreferredChain:     c.GetPreferredChain(),
		Api:                c.GetApi(),
		KeyAlgo:            c.GetKeyAlgo(),
		RenewDays:          c.GetRenewDays(),
		ForceRenew:         c.GetForceRenew(),
		ForceValidation:    c.GetForceValidation(),
		PrivateKeyRenew:    c.GetPrivateKeyRenew(),
		PrivateKeyRollover: c.GetPrivateKeyRollover(),
		ChallengeType:      c.GetChallengeType(),
		WellKnownDir:       c.GetWellKnownDir(),
		AlpnDir:            c.GetAlpnDir(),
		HookChain:          c.GetHookChain(),
		OcspMustStaple:     c.GetOcspMustStaple(),
		OcspFetch:          c.GetOcspFetch(),
		OcspDays:           c.GetOcspDays(),
		NoLock:             c.GetNoLock(),
		KeepGoing:          c.GetKeepGoing(),
		FullChain:          c.GetFullChain(),
		Ocsp:               c.GetOcsp(),
		AutoCleanup:        c.GetAutoCleanup(),
		ContactEmail:       c.GetContactEmail(),
		CurlOpts:           c.GetCurlOpts(),
		ConfigD:            c.GetConfigD(),
	})
}

// NewConfig creates a new Config with default values.
// It initializes all fields with sensible defaults for the dehydrated ACME client.
func NewConfig() *Config {
	return &Config{
		pb.DehydratedConfig{
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
		},
	}
}

// DefaultConfig returns a new Config with default values for production use.
// It sets up the configuration for Let's Encrypt v2 API with standard settings.
func DefaultConfig() *Config {
	return &Config{
		pb.DehydratedConfig{
			Group:    "www-data",
			Ca:       "https://acme-v02.api.letsencrypt.org/directory",
			Openssl:  "openssl",
			OcspDays: 5,
			Api:      "v2",
		},
	}
}

// WithBaseDir sets the base directory for the config.
// This is the root directory where all other paths will be resolved relative to.
func (c *Config) WithBaseDir(baseDir string) *Config {
	c.BaseDir = baseDir
	return c
}

// WithConfigFile sets the path to the config file.
// This file will be used to load configuration settings for dehydrated.
func (c *Config) WithConfigFile(configFile string) *Config {
	c.ConfigFile = configFile
	return c
}

// Load loads the configuration from files.
// It reads the config file if specified, resolves all paths to absolute paths,
// and returns the config for method chaining.
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

// findAndSetConfigFile searches for a config file in the base directory.
// It looks for files named "config" or "config.sh" and sets the ConfigFile field
// if one is found.
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

// load loads configuration from a config file if it exists.
// It parses the file line by line, looking for key-value pairs in the format
// KEY=value or export KEY=value.
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

	c.parse(c.ConfigFile)

	// Resolve relative paths
	c.resolvePaths()
}

func (c *Config) parse(path string) {
	// Read config file
	data, err := os.ReadFile(path)
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
}

// ensureAbs converts a relative path to an absolute path.
// If the path is already absolute, it is returned as is.
// Otherwise, it is joined with the base directory.
func (c *Config) ensureAbs(p string) string {
	if !filepath.IsAbs(p) {
		return filepath.Join(c.BaseDir, p)
	}

	return p
}

// resolvePaths converts relative paths to absolute paths.
// It ensures all paths in the configuration are absolute, using the base directory
// as the root for relative paths.
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

func (c *Config) String() string {
	var lines []string

	t := reflect.TypeOf(*c)
	for i := 0; i < t.NumField(); i++ {
		value := reflect.ValueOf(*c).Field(i)
		if value.String() != "" {
			lines = append(lines, fmt.Sprintf("%s=%v", strings.ToUpper(t.Field(i).Name), reflect.ValueOf(*c).Field(i).Interface()))
		}
	}

	return strings.Join(lines, "\n")
}

func (c *Config) DomainSpecificConfig(path string) *Config {
	cfgFile := filepath.Join(c.CertDir, path, "config")
	if _, err := os.Stat(cfgFile); err != nil {
		return c
	}

	domainSpecificConfig := &Config{}
	domainSpecificConfig.parse(cfgFile)

	cfg := *c
	if domainSpecificConfig.KeyAlgo != "" {
		cfg.KeyAlgo = domainSpecificConfig.KeyAlgo
	}
	if domainSpecificConfig.KeySize > 0 {
		cfg.KeySize = domainSpecificConfig.KeySize
	}
	if domainSpecificConfig.ChallengeType != "" {
		cfg.ChallengeType = domainSpecificConfig.ChallengeType
	}

	return &cfg
}

func (c *Config) ToProto() *pb.DehydratedConfig {
	return &pb.DehydratedConfig{
		BaseDir:         c.BaseDir,
		CertDir:         c.CertDir,
		DomainsDir:      c.DomainsDir,
		AccountsDir:     c.AccountsDir,
		ChallengesDir:   c.ChallengesDir,
		DomainsFile:     c.DomainsFile,
		Ca:              c.Ca,
		OldCa:           c.OldCa,
		RenewDays:       c.RenewDays,
		KeySize:         c.KeySize,
		KeyAlgo:         c.KeyAlgo,
		ChallengeType:   c.ChallengeType,
		WellKnownDir:    c.WellKnownDir,
		LockFile:        c.LockFile,
		Openssl:         c.Openssl,
		PrivateKeyRenew: c.PrivateKeyRenew,
		HookChain:       c.HookChain,
		OcspDays:        c.OcspDays,
		ChainCache:      c.ChainCache,
		Api:             c.Api,
		AcceptTerms:     c.AcceptTerms,
		Ipv4:            c.Ipv4,
		Ipv6:            c.Ipv6,
		PreferredChain:  c.PreferredChain,
		HookScript:      c.HookScript,
		OpensslConfig:   c.OpensslConfig,
		ConfigD:         c.ConfigD,
	}
}
