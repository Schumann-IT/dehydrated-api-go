package dehydrated

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig verifies that the DefaultConfig function returns a Config
// with the expected default values for production use.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.Equal(t, "www-data", config.Group, "Expected Group to be www-data")
	require.Equal(t, "https://acme-v02.api.letsencrypt.org/directory", config.Ca, "Expected Ca to be https://acme-v02.api.letsencrypt.org/directory")
	require.Equal(t, "openssl", config.Openssl, "Expected Openssl to be openssl")
	require.Equal(t, 5, int(config.OcspDays), "Expected OcspDays to be 5")
	require.Equal(t, "v2", config.Api, "Expected Api to be v2")
}

// TestLoadConfig tests loading configuration from a config file.
// It creates a temporary config file with various settings and verifies
// that they are correctly loaded into the Config struct.
//
//nolint:funlen // This test is intentionally long to cover many config options.
func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	baseDir := "/test/base"

	// Create a test config file
	configContent := `# Test config file
BASEDIR=/test/base
CERTDIR=/etc/certs/override
DOMAINSD=domains/override
ACCOUNTDIR=accounts/override
CHALLENGEDIR=challenges/override
DOMAINS_TXT=domains.txt
HOOK=hook.sh
CA=letsencrypt
OLDCA=https://acme-v01.api.letsencrypt.org/directory
ACCEPT_TERMS=yes
IPV4=yes
IPV6=no
PREFERRED_CHAINS=ISRG
KEY_ALGO=prime256v1
KEY_SIZE=2048
RENEW_DAYS=45
FORCE_RENEW=yes
FORCE_VALIDATION=no
CHALLENGETYPE=dns-01
WELLKNOWN=/var/www/dehydrated
ALPNCERTDIR=/var/www/alpn
LOCKFILE=dehydrated.lock
NO_LOCK=no
KEEP_GOING=yes
FULL_CHAIN=yes
OCSP=yes
OPENSSL=/usr/bin/openssl
OPENSSL_CONFIG=/etc/ssl/openssl.cnf
PRIVATE_KEY_RENEW=yes
PRIVATE_KEY_ROLLOVER=no
HOOK_CHAIN=yes
OCSP_MUST_STAPLE=yes
OCSP_FETCH=yes
OCSP_DAYS=10
AUTO_CLEANUP=yes
CONTACT_EMAIL=test@example.com
CURL_OPTS=-k
CONFIG_D=/etc/dehydrated/conf.d
CHAIN_CACHE=chains
API=v2
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config using the builder pattern
	cfg := NewConfig().
		WithConfigFile(configPath).
		Load()

	// Test that values from config.sh override config file values
	if cfg.BaseDir != baseDir {
		t.Errorf("Expected BaseDir to be %s, got %s", baseDir, cfg.BaseDir)
	}
	if cfg.CertDir != "/etc/certs/override" {
		t.Errorf("Expected CertDir to be /etc/certs/override, got %s", cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join(baseDir, "domains", "override") {
		t.Errorf("Expected DomainsDir to be %s, got %s", filepath.Join(baseDir, "domains", "override"), cfg.DomainsDir)
	}
	if cfg.KeyAlgo != "prime256v1" {
		t.Errorf("Expected KeyAlgo to be prime256v1, got %s", cfg.KeyAlgo)
	}
	if cfg.RenewDays != 45 {
		t.Errorf("Expected RenewDays to be 45, got %d", cfg.RenewDays)
	}
	if cfg.ChallengeType != "dns-01" {
		t.Errorf("Expected ChallengeType to be dns-01, got %s", cfg.ChallengeType)
	}

	if cfg.AccountsDir != filepath.Join(baseDir, "accounts", "override") {
		t.Errorf("Expected AccountsDir to be %s, got %s", filepath.Join(baseDir, "accounts"), cfg.AccountsDir)
	}
	if cfg.ChallengesDir != filepath.Join(baseDir, "challenges", "override") {
		t.Errorf("Expected ChallengesDir to be %s, got %s", filepath.Join(baseDir, "challenges"), cfg.ChallengesDir)
	}
	if cfg.DomainsFile != filepath.Join(baseDir, "domains.txt") {
		t.Errorf("Expected DomainsFile to be %s, got %s", filepath.Join(baseDir, "domains.txt\""), cfg.DomainsFile)
	}
	if cfg.HookScript != filepath.Join(baseDir, "hook.sh") {
		t.Errorf("Expected HookScript to be %s, got %s", filepath.Join(baseDir, "hook.sh"), cfg.HookScript)
	}
	if cfg.Ca != "letsencrypt" {
		t.Errorf("Expected Ca to be letsencrypt, got %s", cfg.Ca)
	}
	if cfg.OldCa != "https://acme-v01.api.letsencrypt.org/directory" {
		t.Errorf("Expected OldCa to be https://acme-v01.api.letsencrypt.org/directory, got %s", cfg.OldCa)
	}
	if !cfg.AcceptTerms {
		t.Error("Expected AcceptTerms to be true")
	}
	if !cfg.Ipv4 {
		t.Error("Expected Ipv4 to be true")
	}
	if cfg.Ipv6 {
		t.Error("Expected Ipv6 to be false")
	}
	if cfg.PreferredChain != "ISRG" {
		t.Errorf("Expected PreferredChain to be ISRG, got %s", cfg.PreferredChain)
	}
	if cfg.KeySize != 2048 {
		t.Errorf("Expected KeySize to be 2048, got %d", cfg.KeySize)
	}
	if !cfg.ForceRenew {
		t.Error("Expected ForceRenew to be true")
	}
	if cfg.ForceValidation {
		t.Error("Expected ForceValidation to be false")
	}
	if cfg.WellKnownDir != "/var/www/dehydrated" {
		t.Errorf("Expected WellKnownDir to be /var/www/dehydrated, got %s", cfg.WellKnownDir)
	}
	if cfg.AlpnDir != "/var/www/alpn" {
		t.Errorf("Expected AlpnDir to be /var/www/alpn, got %s", cfg.AlpnDir)
	}
	if cfg.LockFile != filepath.Join(baseDir, "dehydrated.lock") {
		t.Errorf("Expected LockFile to be %s, got %s", filepath.Join(baseDir, "dehydrated.lock"), cfg.LockFile)
	}
	if cfg.NoLock {
		t.Error("Expected NoLock to be false")
	}
	if !cfg.KeepGoing {
		t.Error("Expected KeepGoing to be true")
	}
	if !cfg.FullChain {
		t.Error("Expected FullChain to be true")
	}
	if !cfg.Ocsp {
		t.Error("Expected Ocsp to be true")
	}
	if cfg.Openssl != "/usr/bin/openssl" {
		t.Errorf("Expected Openssl to be /usr/bin/openssl, got %s", cfg.Openssl)
	}
	if cfg.OpensslConfig != "/etc/ssl/openssl.cnf" {
		t.Errorf("Expected OpensslConfig to be /etc/ssl/openssl.cnf, got %s", cfg.OpensslConfig)
	}
	if !cfg.PrivateKeyRenew {
		t.Error("Expected PrivateKeyRenew to be true")
	}
	if cfg.PrivateKeyRollover {
		t.Error("Expected PrivateKeyRollover to be false")
	}
	if !cfg.HookChain {
		t.Error("Expected HookChain to be true")
	}
	if !cfg.OcspMustStaple {
		t.Error("Expected OcspMustStaple to be true")
	}
	if !cfg.OcspFetch {
		t.Error("Expected OcspFetch to be true")
	}
	if cfg.OcspDays != 10 {
		t.Errorf("Expected OcspDays to be 10, got %d", cfg.OcspDays)
	}
	if !cfg.AutoCleanup {
		t.Error("Expected AutoCleanup to be true")
	}
	if cfg.ContactEmail != "test@example.com" {
		t.Errorf("Expected ContactEmail to be test@example.com, got %s", cfg.ContactEmail)
	}
	if cfg.CurlOpts != "-k" {
		t.Errorf("Expected CurlOpts to be -k, got %s", cfg.CurlOpts)
	}
	if cfg.ConfigD != "/etc/dehydrated/conf.d" {
		t.Errorf("Expected ConfigD to be /etc/dehydrated/conf.d, got %s", cfg.ConfigD)
	}
	if cfg.ChainCache != filepath.Join(baseDir, "chains") {
		t.Errorf("Expected ChainCache to be %s, got %s", filepath.Join(baseDir, "chains"), cfg.ChainCache)
	}
	if cfg.Api != "v2" {
		t.Errorf("Expected Api to be v2, got %s", cfg.Api)
	}
}

// TestLoadConfigFromShellScript tests loading configuration from a shell script.
// It verifies that the script is correctly parsed and the settings are loaded.
//
//nolint:funlen // This test is intentionally long to cover many config options.
func TestLoadConfigFromShellScript(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	// Create a test config file
	configContent := `#!/bin/sh

# Test config file
BASEDIR=/test/base
CERTDIR=/etc/certs/override
DOMAINSD=domains/override
ACCOUNTDIR=accounts/override
CHALLENGEDIR=challenges/override
DOMAINS_TXT=domains.txt
HOOK=hook.sh
CA=https://acme-v02.api.letsencrypt.org/directory
ACCEPT_TERMS=yes
IPV4=yes
IPV6=no
PREFERRED_CHAINS=ISRG
KEY_ALGO=prime256v1
KEY_SIZE=2048
RENEW_DAYS=45
WELLKNOWN=/var/www/dehydrated
ALPNCERTDIR=/etc/alpn
LOCKFILE=/var/lock/dehydrated.lock
OCSP=yes
OPENSSL=/usr/bin/openssl
OPENSSL_CONFIG=/etc/ssl/openssl.cnf
PRIVATE_KEY_RENEW=yes
HOOK_CHAIN=yes
OCSP_MUST_STAPLE=yes
API=v2
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg := NewConfig().WithBaseDir("/test/base").WithConfigFile(configPath).Load()

	// Test loaded values
	if cfg.BaseDir != "/test/base" {
		t.Errorf("Expected BaseDir to be /test/base, got %s", cfg.BaseDir)
	}
	if cfg.CertDir != "/etc/certs/override" {
		t.Errorf("Expected CertDir to be /etc/certs/override, got %s", cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join("/test", "base", "domains", "override") {
		t.Errorf("Expected DomainsDir to be %s, got %s", filepath.Join("/test", "base", "domains", "override"), cfg.DomainsDir)
	}
	if cfg.AccountsDir != filepath.Join("/test", "base", "accounts", "override") {
		t.Errorf("Expected AccountsDir to be %s, got %s", filepath.Join("/test", "base", "accounts", "override"), cfg.AccountsDir)
	}
	if cfg.ChallengesDir != filepath.Join("/test", "base", "challenges", "override") {
		t.Errorf("Expected ChallengesDir to be %s, got %s", filepath.Join("/test", "base", "challenges", "override"), cfg.ChallengesDir)
	}
	if cfg.DomainsFile != filepath.Join("/test", "base", "domains.txt") {
		t.Errorf("Expected DomainsFile to be %s, got %s", filepath.Join("/test", "base", "domains.txt"), cfg.DomainsFile)
	}
	if cfg.HookScript != "/test/base/hook.sh" {
		t.Errorf("Expected HookScript to be /test/base/hook.sh, got %s", cfg.HookScript)
	}
	if cfg.Ca != "https://acme-v02.api.letsencrypt.org/directory" {
		t.Errorf("Expected Ca to be https://acme-v02.api.letsencrypt.org/directory, got %s", cfg.Ca)
	}
	if !cfg.AcceptTerms {
		t.Error("Expected AcceptTerms to be true")
	}
	if !cfg.Ipv4 {
		t.Error("Expected Ipv4 to be true")
	}
	if cfg.Ipv6 {
		t.Error("Expected Ipv6 to be false")
	}
	if cfg.PreferredChain != "ISRG" {
		t.Errorf("Expected PreferredChain to be ISRG, got %s", cfg.PreferredChain)
	}
	if cfg.KeyAlgo != "prime256v1" {
		t.Errorf("Expected KeyAlgo to be prime256v1, got %s", cfg.KeyAlgo)
	}
	if cfg.KeySize != 2048 {
		t.Errorf("Expected KeySize to be 2048, got %d", cfg.KeySize)
	}
	if cfg.RenewDays != 45 {
		t.Errorf("Expected RenewDays to be 45, got %d", cfg.RenewDays)
	}
	if cfg.WellKnownDir != "/var/www/dehydrated" {
		t.Errorf("Expected WellKnownDir to be /var/www/dehydrated, got %s", cfg.WellKnownDir)
	}
	if cfg.AlpnDir != "/etc/alpn" {
		t.Errorf("Expected AlpnDir to be /etc/alpn, got %s", cfg.AlpnDir)
	}
	if cfg.LockFile != "/var/lock/dehydrated.lock" {
		t.Errorf("Expected LockFile to be /var/lock/dehydrated.lock, got %s", cfg.LockFile)
	}
	if !cfg.Ocsp {
		t.Error("Expected Ocsp to be true")
	}
	if cfg.Openssl != "/usr/bin/openssl" {
		t.Errorf("Expected Openssl to be /usr/bin/openssl, got %s", cfg.Openssl)
	}
	if cfg.OpensslConfig != "/etc/ssl/openssl.cnf" {
		t.Errorf("Expected OpensslConfig to be /etc/ssl/openssl.cnf, got %s", cfg.OpensslConfig)
	}
	if !cfg.PrivateKeyRenew {
		t.Error("Expected PrivateKeyRenew to be true")
	}
	if !cfg.HookChain {
		t.Error("Expected HookChain to be true")
	}
	if !cfg.OcspMustStaple {
		t.Error("Expected OcspMustStaple to be true")
	}
	if cfg.Api != "v2" {
		t.Errorf("Expected Api to be v2, got %s", cfg.Api)
	}
}

// TestLoadConfigWithShellScriptFromFixtures tests loading configuration from
// a shell script in the fixtures directory. It verifies that the script is
// correctly parsed and the settings are loaded.
func TestLoadConfigWithShellScriptFromFixtures(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.sh")

	// Create a test config file
	configContent := `#!/bin/bash

# Which public key algorithm should be used? Supported: rsa, prime256v1 and secp384r1
KEY_ALGO=prime256v1

KEYSIZE="4096"
PRIVATE_KEY_RENEW="yes"

#USE BELOW CA FOR TESTING OTHERWISE YOU MIGHT GET BANNED FROM LE https://community.letsencrypt.org/t/rate-limits-for-lets-encrypt/6769
#CA="https://acme-staging-v02.api.letsencrypt.org/directory"

# Minimum days before expiration to automatically renew certificate (default: 30)
# 90 Tage ist ein Cert gültig
RENEW_DAYS="60"

#HOOK="/home/le-user/dns-challenge/"
HOOK_CHAIN="no"

WELLKNOWN="/home/le-user/dns-challenge/www-temp"
# E-mail to use during the registration (default: <unset>)
CONTACT_EMAIL="webmaster@hansemerkur.de"

#Challenge Type
#Change to dns-01 for DNS challenge
CHALLENGETYPE="dns-01"

# OCSP Stapeling
# Option to add CSR-flag indicating OCSP stapling to be mandatory (default: no)
# OCSP_MUST_STAPLE="yes"
# Fetch OCSP responses (default: no)
#OCSP_FETCH="no"
# OCSP refresh interval (default: 5 days)
#OCSP_DAYS=5
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	abs, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	cfg := NewConfig().WithBaseDir(tmpDir).Load()

	if cfg.BaseDir != abs {
		t.Errorf("Expected BaseDir to be %s, got %s", abs, cfg.BaseDir)
	}
	if cfg.CertDir != filepath.Join(abs, "certs") {
		t.Errorf("Expected CertDir to be %s, got %s", filepath.Join(abs, "certs"), cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join(abs, "domains") {
		t.Errorf("Expected DomainsDir to be %s, got %s", filepath.Join(abs, "domains"), cfg.DomainsDir)
	}
	if cfg.KeyAlgo != "prime256v1" {
		t.Errorf("Expected KeyAlgo to be prime256v1, got %s", cfg.KeyAlgo)
	}
	if cfg.RenewDays != 60 {
		t.Errorf("Expected RenewDays to be 60, got %d", cfg.RenewDays)
	}
	if cfg.ChallengeType != "dns-01" {
		t.Errorf("Expected ChallengeType to be dns-01, got %s", cfg.ChallengeType)
	}
	if cfg.WellKnownDir != "/home/le-user/dns-challenge/www-temp" {
		t.Errorf("Expected WellKnownDir to be /home/le-user/dns-challenge/www-temp, got %s", cfg.WellKnownDir)
	}
}

// TestConfig tests various Config methods including WithBaseDir, WithConfigFile,
// and Load. It verifies that the methods work correctly and return the expected
// Config instance.
func TestConfig(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		cfg := DefaultConfig()

		if cfg.Ca != "https://acme-v02.api.letsencrypt.org/directory" {
			t.Errorf("Expected Ca to be https://acme-v02.api.letsencrypt.org/directory, got %s", cfg.Ca)
		}

		if cfg.OldCa != "" {
			t.Errorf("Expected OldCa to be empty, got %s", cfg.OldCa)
		}

		if cfg.ContactEmail != "" {
			t.Errorf("Expected ContactEmail to be empty, got %s", cfg.ContactEmail)
		}

		if cfg.Openssl != "openssl" {
			t.Errorf("Expected Openssl to be openssl, got %s", cfg.Openssl)
		}

		if cfg.Group != "www-data" {
			t.Errorf("Expected Group to be www-data, got %s", cfg.Group)
		}

		if cfg.OcspDays != 5 {
			t.Errorf("Expected OcspDays to be 5, got %d", cfg.OcspDays)
		}

		if cfg.Api != "v2" {
			t.Errorf("Expected Api to be v2, got %s", cfg.Api)
		}
	})

	t.Run("LoadFromFile", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		// Create a test config file
		configContent := `#!/bin/sh
CA="https://acme-staging-v02.api.letsencrypt.org/directory"
OLDCA="https://acme-v01.api.letsencrypt.org/directory"
CONTACT_EMAIL="admin@example.com"
OPENSSL="/usr/bin/openssl"
GROUP="www-data"
OCSP_DAYS=10
API="v2"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Load the config
		cfg := &Config{}
		cfg = cfg.WithConfigFile(configPath).Load()

		// Verify values
		if cfg.Ca != "https://acme-staging-v02.api.letsencrypt.org/directory" {
			t.Errorf("Expected Ca to be https://acme-staging-v02.api.letsencrypt.org/directory, got %s", cfg.Ca)
		}

		if cfg.OldCa != "https://acme-v01.api.letsencrypt.org/directory" {
			t.Errorf("Expected OldCa to be https://acme-v01.api.letsencrypt.org/directory, got %s", cfg.OldCa)
		}

		if cfg.ContactEmail != "admin@example.com" {
			t.Errorf("Expected ContactEmail to be admin@example.com, got %s", cfg.ContactEmail)
		}

		if cfg.Openssl != "/usr/bin/openssl" {
			t.Errorf("Expected Openssl to be /usr/bin/openssl, got %s", cfg.Openssl)
		}

		if cfg.Group != "www-data" {
			t.Errorf("Expected Group to be www-data, got %s", cfg.Group)
		}

		if cfg.OcspDays != 10 {
			t.Errorf("Expected OcspDays to be 10, got %d", cfg.OcspDays)
		}

		if cfg.Api != "v2" {
			t.Errorf("Expected Api to be v2, got %s", cfg.Api)
		}
	})
}

func TestConfig_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "all fields set",
			config: &Config{
				DehydratedConfig: pb.DehydratedConfig{
					User:               "testuser",
					Group:              "testgroup",
					BaseDir:            "/test/base",
					CertDir:            "/test/certs",
					DomainsDir:         "/test/domains",
					AccountsDir:        "/test/accounts",
					ChallengesDir:      "/test/challenges",
					ChainCache:         "/test/chains",
					DomainsFile:        "domains.txt",
					ConfigFile:         "config",
					HookScript:         "hook.sh",
					LockFile:           "lock",
					OpensslConfig:      "openssl.cnf",
					Openssl:            "openssl",
					KeySize:            4096,
					Ca:                 "letsencrypt",
					OldCa:              "old-ca",
					AcceptTerms:        true,
					Ipv4:               true,
					Ipv6:               true,
					PreferredChain:     "chain1",
					Api:                "v2",
					KeyAlgo:            "rsa",
					RenewDays:          30,
					ForceRenew:         true,
					ForceValidation:    true,
					PrivateKeyRenew:    true,
					PrivateKeyRollover: true,
					ChallengeType:      "http-01",
					WellKnownDir:       "/test/well-known",
					AlpnDir:            "/test/alpn",
					HookChain:          true,
					OcspMustStaple:     true,
					OcspFetch:          true,
					OcspDays:           7,
					NoLock:             true,
					KeepGoing:          true,
					FullChain:          true,
					Ocsp:               true,
					AutoCleanup:        true,
					ContactEmail:       "test@example.com",
					CurlOpts:           "--test",
					ConfigD:            "/test/config.d",
				},
			},
			expected: `{
				"user": "testuser",
				"group": "testgroup",
				"base_dir": "/test/base",
				"cert_dir": "/test/certs",
				"domains_dir": "/test/domains",
				"accounts_dir": "/test/accounts",
				"challenges_dir": "/test/challenges",
				"chain_cache": "/test/chains",
				"domains_file": "domains.txt",
				"config_file": "config",
				"hook_script": "hook.sh",
				"lock_file": "lock",
				"openssl_config": "openssl.cnf",
				"openssl": "openssl",
				"key_size": 4096,
				"ca": "letsencrypt",
				"old_ca": "old-ca",
				"accept_terms": true,
				"ipv4": true,
				"ipv6": true,
				"preferred_chain": "chain1",
				"api": "v2",
				"key_algo": "rsa",
				"renew_days": 30,
				"force_renew": true,
				"force_validation": true,
				"private_key_renew": true,
				"private_key_rollover": true,
				"challenge_type": "http-01",
				"well_known_dir": "/test/well-known",
				"alpn_dir": "/test/alpn",
				"hook_chain": true,
				"ocsp_must_staple": true,
				"ocsp_fetch": true,
				"ocsp_days": 7,
				"no_lock": true,
				"keep_going": true,
				"full_chain": true,
				"ocsp": true,
				"auto_cleanup": true,
				"contact_email": "test@example.com",
				"curl_opts": "--test",
				"config_d": "/test/config.d"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the config
			actual, err := json.Marshal(tt.config)
			require.NoError(t, err)

			// Compare JSON objects (ignoring whitespace)
			var actualJSON, expectedJSON any
			err = json.Unmarshal(actual, &actualJSON)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(tt.expected), &expectedJSON)
			require.NoError(t, err)

			require.Equal(t, expectedJSON, actualJSON)
		})
	}
}
