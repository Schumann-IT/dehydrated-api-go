package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
)

func createTestCertificate(t *testing.T, certDir, domain string) {
	// Create domain directory
	domainDir := filepath.Join(certDir, domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatalf("Failed to create domain directory: %v", err)
	}

	// Generate CA key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CA key: %v", err)
	}

	// Create CA template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	// Generate domain key
	domainKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate domain key: %v", err)
	}

	// Create domain certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	// Create domain certificate
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &domainKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Write certificate to file
	certFile := filepath.Join(domainDir, "cert.pem")
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert.pem: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("Failed to write cert.pem: %v", err)
	}
}

func TestCertsPlugin(t *testing.T) {
	// Create temporary directory for test certificates
	tmpDir := t.TempDir()
	certDir := filepath.Join(tmpDir, "certs")

	// Create test domain and certificate
	domain := "example.com"
	createTestCertificate(t, certDir, domain)

	// Create config with absolute path
	cfg := config.NewConfig()
	cfg.BaseDir = tmpDir
	cfg.CertDir = certDir

	t.Run("Initialize", func(t *testing.T) {
		plugin := New()
		if err := plugin.Initialize(cfg); err != nil {
			t.Errorf("Failed to initialize plugin: %v", err)
		}
		if plugin.certDir != certDir {
			t.Errorf("Expected cert dir %s, got %s", certDir, plugin.certDir)
		}
	})

	t.Run("EnrichDomainEntry", func(t *testing.T) {
		plugin := New()
		if err := plugin.Initialize(cfg); err != nil {
			t.Fatalf("Failed to initialize plugin: %v", err)
		}

		entry := &model.DomainEntry{
			Domain: domain,
		}
		if err := plugin.EnrichDomainEntry(entry); err != nil {
			t.Errorf("Failed to enrich domain entry: %v", err)
		}

		certInfo, ok := entry.Metadata["cert"].(*CertInfo)
		if !ok {
			t.Fatal("Expected cert info in metadata")
		}
		if !certInfo.IsValid {
			t.Error("Expected certificate to be valid")
		}
		if certInfo.Issuer != "Test CA" {
			t.Errorf("Expected issuer Test CA, got %s", certInfo.Issuer)
		}
	})

	t.Run("EnrichDomainEntryNoCert", func(t *testing.T) {
		plugin := New()
		if err := plugin.Initialize(cfg); err != nil {
			t.Fatalf("Failed to initialize plugin: %v", err)
		}

		entry := &model.DomainEntry{
			Domain: "nonexistent.com",
		}
		if err := plugin.EnrichDomainEntry(entry); err != nil {
			t.Errorf("Failed to enrich domain entry: %v", err)
		}

		certInfo, ok := entry.Metadata["cert"].(*CertInfo)
		if !ok {
			t.Fatal("Expected cert info in metadata")
		}
		if certInfo.IsValid {
			t.Error("Expected certificate to be invalid")
		}
		if certInfo.Error == "" {
			t.Error("Expected error message")
		}
	})
}
