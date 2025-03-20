package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test default values
	if cfg.BaseDir != "/etc/dehydrated" {
		t.Errorf("Expected BaseDir to be /etc/dehydrated, got %s", cfg.BaseDir)
	}
	if cfg.CertDir != "certs" {
		t.Errorf("Expected CertDir to be certs, got %s", cfg.CertDir)
	}
	if cfg.DomainsDir != "domains" {
		t.Errorf("Expected DomainsDir to be domains, got %s", cfg.DomainsDir)
	}
	if cfg.AccountsDir != "accounts" {
		t.Errorf("Expected AccountsDir to be accounts, got %s", cfg.AccountsDir)
	}
	if cfg.ChallengesDir != "acme-challenges" {
		t.Errorf("Expected ChallengesDir to be acme-challenges, got %s", cfg.ChallengesDir)
	}
	if cfg.DomainsFile != "domains.txt" {
		t.Errorf("Expected DomainsFile to be domains.txt, got %s", cfg.DomainsFile)
	}
	if cfg.RenewDays != 30 {
		t.Errorf("Expected RenewDays to be 30, got %d", cfg.RenewDays)
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
	if cfg.LockFile != "dehydrated.lock" {
		t.Errorf("Expected LockFile to be dehydrated.lock, got %s", cfg.LockFile)
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
CA=https://acme-v02.api.letsencrypt.org/directory
ACCEPT_TERMS=yes
IPV4=yes
IPV6=no
PREFERRED_CHAINS=ISRG
KEY_ALGO=prime256v1
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
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

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

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	// Create a test config file
	err := os.WriteFile(configPath, []byte("# Test config file"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Temporarily change the working directory to the temp dir
	oldPWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldPWD)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test finding the config file
	found, err := FindConfigFile()
	if err != nil {
		t.Fatalf("Failed to find config file: %v", err)
	}

	// Evaluate symlinks in both paths
	foundEval, err := filepath.EvalSymlinks(found)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks in found path: %v", err)
	}
	configPathEval, err := filepath.EvalSymlinks(configPath)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks in config path: %v", err)
	}

	// Compare the resolved paths
	if foundEval != configPathEval {
		t.Errorf("Expected config file at %s, got %s", configPathEval, foundEval)
	}
}
