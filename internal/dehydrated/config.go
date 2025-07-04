// Package dehydrated provides functionality for working with the dehydrated ACME client.
// It includes configuration management, path resolution, and integration with the dehydrated script.

package dehydrated

import (
	"errors"
	"fmt"
	"math"
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
	if filepath.IsAbs(configFile) {
		c.ConfigFile = configFile
	} else {
		// If the config file is relative, resolve it against the base directory
		c.ConfigFile = filepath.Join(c.BaseDir, configFile)
	}

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

//nolint:gocyclo,funlen // this function needs refactoring @TODO strip down the number of fields
func (c *Config) SetValue(key, value string) {
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
		val, err := toInt32(value)
		if err != nil {
			break
		}
		c.KeySize = val
	case "RENEW_DAYS":
		val, err := toInt32(value)
		if err != nil {
			break
		}
		c.RenewDays = val
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
		val, err := toInt32(value)
		if err != nil {
			break
		}
		c.OcspDays = val
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

func (c *Config) parse(path string) {
	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Parse config file
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		key, value, err := trimLine(line)
		if err != nil {
			continue
		}

		c.SetValue(key, value)
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
	c.ChainCache = c.ensureAbs(c.ChainCache)

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

// String returns a string representation of the Config.
func (c *Config) String() string {
	var lines []string

	//nolint: govet // We use reflection to iterate over fields
	t := reflect.TypeOf(*c)
	for i := 0; i < t.NumField(); i++ {
		//nolint: govet // We use reflection to iterate over fields
		value := reflect.ValueOf(*c).Field(i)
		if value.String() != "" {
			//nolint: govet // We use reflection to iterate over fields
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

	if domainSpecificConfig.KeyAlgo != "" {
		c.KeyAlgo = domainSpecificConfig.KeyAlgo
	}
	if domainSpecificConfig.KeySize > 0 {
		c.KeySize = domainSpecificConfig.KeySize
	}
	if domainSpecificConfig.ChallengeType != "" {
		c.ChallengeType = domainSpecificConfig.ChallengeType
	}

	return c
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

func trimLine(line string) (string, string, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", errors.New("empty or comment line")
	}

	// Look for export statements
	line = strings.TrimPrefix(line, "export ")

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid line format, expected KEY=VALUE")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Remove quotes if present
	value = strings.Trim(value, "\"'")

	return key, value, nil
}

func toInt32(value string) (int32, error) {
	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt32 || val > math.MaxInt32 {
		return 0, fmt.Errorf("value %d is out of int32 range", val)
	}

	//nolint:gosec // We are converting a string to int32, no security risk here
	return int32(val), nil
}
