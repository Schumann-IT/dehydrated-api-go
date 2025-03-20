package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := NewConfig().WithBaseDir(tmpDir).Load()

	// Test default values
	if cfg.BaseDir != tmpDir {
		t.Errorf("Expected BaseDir to be %s, got %s", tmpDir, cfg.BaseDir)
	}
	if cfg.CertDir != filepath.Join(tmpDir, "certs") {
		t.Errorf("Expected CertDir to be %s, got %s", filepath.Join(tmpDir, "certs"), cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join(tmpDir, "domains") {
		t.Errorf("Expected DomainsDir to be %s, got %s", filepath.Join(tmpDir, "domains"), cfg.DomainsDir)
	}
	if cfg.AccountsDir != filepath.Join(tmpDir, "accounts") {
		t.Errorf("Expected AccountsDir to be %s, got %s", filepath.Join(tmpDir, "accounts"), cfg.DomainsDir)
	}
	if cfg.ChallengesDir != filepath.Join(tmpDir, "acme-challenges") {
		t.Errorf("Expected ChallengesDir to be %s, got %s", filepath.Join(tmpDir, "acme-challenges"), cfg.ChallengesDir)
	}
	if cfg.DomainsFile != filepath.Join(tmpDir, "domains.txt") {
		t.Errorf("Expected DomainsFile to be %s, got %s", filepath.Join(tmpDir, "domains.txt"), cfg.DomainsFile)
	}
	if cfg.CA != "letsencrypt" {
		t.Errorf("Expected CA to be letsencrypt, got %s", cfg.CA)
	}
	if cfg.OldCA != "https://acme-v01.api.letsencrypt.org/directory" {
		t.Errorf("Expected OldCA to be https://acme-v01.api.letsencrypt.org/directory, got %s", cfg.OldCA)
	}
	if cfg.RenewDays != 30 {
		t.Errorf("Expected RenewDays to be 30, got %d", cfg.RenewDays)
	}
	if cfg.KeySize != 4096 {
		t.Errorf("Expected KeySize to be 4096, got %d", cfg.KeySize)
	}
	if cfg.KeyAlgo != "rsa" {
		t.Errorf("Expected KeyAlgo to be rsa, got %s", cfg.KeyAlgo)
	}
	if cfg.ChallengeType != "http-01" {
		t.Errorf("Expected ChallengeType to be http-01, got %s", cfg.ChallengeType)
	}
	if cfg.WellKnownDir != "/var/www/dehydrated" {
		t.Errorf("Expected WellKnownDir to be /var/www/dehydrated, got %s", cfg.WellKnownDir)
	}
	if cfg.LockFile != filepath.Join(tmpDir, "dehydrated.lock") {
		t.Errorf("Expected LockFile to be %s, got %s", filepath.Join(tmpDir, "dehydrated.lock"), cfg.LockFile)
	}
	if cfg.OpenSSL != "openssl" {
		t.Errorf("Expected OpenSSL to be openssl, got %s", cfg.OpenSSL)
	}
	if !cfg.PrivateKeyRenew {
		t.Error("Expected PrivateKeyRenew to be true")
	}
	if cfg.HookChain {
		t.Error("Expected HookChain to be false")
	}
	if cfg.OCSPDays != 5 {
		t.Errorf("Expected OCSPDays to be 5, got %d", cfg.OCSPDays)
	}
	if cfg.ChainCache != "chains" {
		t.Errorf("Expected ChainCache to be chains, got %s", cfg.ChainCache)
	}
	if cfg.API != "auto" {
		t.Errorf("Expected API to be auto, got %s", cfg.API)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	// Create a test config file
	configContent := `# Test config file
BASEDIR=/test/base
CERTDIR=certs
DOMAINSD=domains
ACCOUNTDIR=accounts
CHALLENGEDIR=challenges
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

	// Load the config
	cfg := NewConfig().WithBaseDir(tmpDir).Load()

	// Test loaded values
	if cfg.BaseDir != "/test/base" {
		t.Errorf("Expected BaseDir to be /test/base, got %s", cfg.BaseDir)
	}
	if cfg.CertDir != filepath.Join("/test/base", "certs") {
		t.Errorf("Expected CertDir to be /test/base/certs, got %s", cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join("/test/base", "domains") {
		t.Errorf("Expected DomainsDir to be /test/base/domains, got %s", cfg.DomainsDir)
	}
	if cfg.AccountsDir != filepath.Join("/test/base", "accounts") {
		t.Errorf("Expected AccountsDir to be /test/base/accounts, got %s", cfg.AccountsDir)
	}
	if cfg.ChallengesDir != filepath.Join("/test/base", "challenges") {
		t.Errorf("Expected ChallengesDir to be /test/base/challenges, got %s", cfg.ChallengesDir)
	}
	if cfg.DomainsFile != filepath.Join("/test/base", "domains.txt") {
		t.Errorf("Expected DomainsFile to be /test/base/domains.txt, got %s", cfg.DomainsFile)
	}
	if cfg.HookScript != filepath.Join("/test/base", "hook.sh") {
		t.Errorf("Expected HookScript to be /test/base/hook.sh, got %s", cfg.HookScript)
	}
	if cfg.CA != "letsencrypt" {
		t.Errorf("Expected CA to be letsencrypt, got %s", cfg.CA)
	}
	if cfg.OldCA != "https://acme-v01.api.letsencrypt.org/directory" {
		t.Errorf("Expected OldCA to be https://acme-v01.api.letsencrypt.org/directory, got %s", cfg.OldCA)
	}
	if !cfg.AcceptTerms {
		t.Error("Expected AcceptTerms to be true")
	}
	if !cfg.IPV4 {
		t.Error("Expected IPV4 to be true")
	}
	if cfg.IPV6 {
		t.Error("Expected IPV6 to be false")
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
	if !cfg.ForceRenew {
		t.Error("Expected ForceRenew to be true")
	}
	if cfg.ForceValidation {
		t.Error("Expected ForceValidation to be false")
	}
	if cfg.ChallengeType != "dns-01" {
		t.Errorf("Expected ChallengeType to be dns-01, got %s", cfg.ChallengeType)
	}
	if cfg.WellKnownDir != "/var/www/dehydrated" {
		t.Errorf("Expected WellKnownDir to be /var/www/dehydrated, got %s", cfg.WellKnownDir)
	}
	if cfg.ALPNDir != "/var/www/alpn" {
		t.Errorf("Expected ALPNDir to be /var/www/alpn, got %s", cfg.ALPNDir)
	}
	if cfg.LockFile != filepath.Join("/test/base", "dehydrated.lock") {
		t.Errorf("Expected LockFile to be /test/base/dehydrated.lock, got %s", cfg.LockFile)
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
	if !cfg.OCSP {
		t.Error("Expected OCSP to be true")
	}
	if cfg.OpenSSL != "/usr/bin/openssl" {
		t.Errorf("Expected OpenSSL to be /usr/bin/openssl, got %s", cfg.OpenSSL)
	}
	if cfg.OpenSSLConfig != "/etc/ssl/openssl.cnf" {
		t.Errorf("Expected OpenSSLConfig to be /etc/ssl/openssl.cnf, got %s", cfg.OpenSSLConfig)
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
	if !cfg.OCSPMustStaple {
		t.Error("Expected OCSPMustStaple to be true")
	}
	if !cfg.OCSPFetch {
		t.Error("Expected OCSPFetch to be true")
	}
	if cfg.OCSPDays != 10 {
		t.Errorf("Expected OCSPDays to be 10, got %d", cfg.OCSPDays)
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
	if cfg.ChainCache != "chains" {
		t.Errorf("Expected ChainCache to be chains, got %s", cfg.ChainCache)
	}
	if cfg.API != "v2" {
		t.Errorf("Expected API to be v2, got %s", cfg.API)
	}
}

func TestLoadConfigFromShellScript(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.sh")

	// Create a test config.sh file
	configContent := `#!/bin/bash

# Test config.sh file
export BASEDIR=/test/base
export CERTDIR=certs
export DOMAINSD=domains
export ACCOUNTDIR=accounts
export CHALLENGEDIR=challenges
export DOMAINS_TXT=domains.txt
export HOOK=hook.sh
export CA=https://acme-v02.api.letsencrypt.org/directory
export ACCEPT_TERMS=yes
export IPV4=yes
export IPV6=no
export PREFERRED_CHAINS=ISRG
export KEY_ALGO=prime256v1
export RENEW_DAYS=45
export FORCE_RENEW=yes
export FORCE_VALIDATION=no
export CHALLENGETYPE=dns-01
export WELLKNOWN=/var/www/dehydrated
export ALPNCERTDIR=/var/www/alpn
export LOCKFILE=dehydrated.lock
export NO_LOCK=no
export KEEP_GOING=yes
export FULL_CHAIN=yes
export OCSP=yes
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config.sh file: %v", err)
	}

	cfg := NewConfig().WithBaseDir(tmpDir).Load()

	// Test loaded values
	if cfg.BaseDir != "/test/base" {
		t.Errorf("Expected BaseDir to be /test/base, got %s", cfg.BaseDir)
	}
	if cfg.CertDir != filepath.Join("/test/base", "certs") {
		t.Errorf("Expected CertDir to be /test/base/certs, got %s", cfg.CertDir)
	}
	if cfg.DomainsDir != filepath.Join("/test/base", "domains") {
		t.Errorf("Expected DomainsDir to be /test/base/domains, got %s", cfg.DomainsDir)
	}
	if cfg.AccountsDir != filepath.Join("/test/base", "accounts") {
		t.Errorf("Expected AccountsDir to be /test/base/accounts, got %s", cfg.AccountsDir)
	}
	if cfg.ChallengesDir != filepath.Join("/test/base", "challenges") {
		t.Errorf("Expected ChallengesDir to be /test/base/challenges, got %s", cfg.ChallengesDir)
	}
	if cfg.DomainsFile != filepath.Join("/test/base", "domains.txt") {
		t.Errorf("Expected DomainsFile to be /test/base/domains.txt, got %s", cfg.DomainsFile)
	}
	if cfg.HookScript != filepath.Join("/test/base", "hook.sh") {
		t.Errorf("Expected HookScript to be /test/base/hook.sh, got %s", cfg.HookScript)
	}
	if cfg.CA != "https://acme-v02.api.letsencrypt.org/directory" {
		t.Errorf("Expected CA to be https://acme-v02.api.letsencrypt.org/directory, got %s", cfg.CA)
	}
	if !cfg.AcceptTerms {
		t.Error("Expected AcceptTerms to be true")
	}
	if !cfg.IPV4 {
		t.Error("Expected IPV4 to be true")
	}
	if cfg.IPV6 {
		t.Error("Expected IPV6 to be false")
	}
	if cfg.PreferredChain != "ISRG" {
		t.Errorf("Expected PreferredChain to be ISRG, got %s", cfg.PreferredChain)
	}
	if cfg.KeyAlgo != "prime256v1" {
		t.Errorf("Expected KeyAlgo to be prime256v1, got %s", cfg.KeyAlgo)
	}
	if cfg.RenewDays != 45 {
		t.Errorf("Expected RenewDays to be 45, got %d", cfg.RenewDays)
	}
	if !cfg.ForceRenew {
		t.Error("Expected ForceRenew to be true")
	}
	if cfg.ForceValidation {
		t.Error("Expected ForceValidation to be false")
	}
	if cfg.ChallengeType != "dns-01" {
		t.Errorf("Expected ChallengeType to be dns-01, got %s", cfg.ChallengeType)
	}
	if cfg.WellKnownDir != "/var/www/dehydrated" {
		t.Errorf("Expected WellKnownDir to be /var/www/dehydrated, got %s", cfg.WellKnownDir)
	}
	if cfg.ALPNDir != "/var/www/alpn" {
		t.Errorf("Expected ALPNDir to be /var/www/alpn, got %s", cfg.ALPNDir)
	}
	if cfg.LockFile != filepath.Join("/test/base", "dehydrated.lock") {
		t.Errorf("Expected LockFile to be /test/base/dehydrated.lock, got %s", cfg.LockFile)
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
	if !cfg.OCSP {
		t.Error("Expected OCSP to be true")
	}
}

func TestLoadConfigWithShellScriptFromFixtures(t *testing.T) {
	abs, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}

	cfg := NewConfig().WithBaseDir("testdata").Load()

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
