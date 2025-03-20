package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
)

// CertInfo represents certificate information
type CertInfo struct {
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	SerialNumber string    `json:"serial_number"`
	Issuer       string    `json:"issuer"`
	IsValid      bool      `json:"is_valid"`
	Error        string    `json:"error,omitempty"`
}

// Plugin implements the plugin interface for certificate information
type Plugin struct {
	certDir string
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "certs"
}

// Initialize sets up the plugin
func (p *Plugin) Initialize(cfg *config.Config) error {
	p.certDir = cfg.CertDir
	return nil
}

// EnrichDomainEntry adds certificate information to the domain entry
func (p *Plugin) EnrichDomainEntry(entry *model.DomainEntry) error {
	certInfo, err := p.getCertInfo(entry.Domain)
	if err != nil {
		// Don't fail if we can't read the certificate, just add error info
		certInfo = &CertInfo{
			IsValid: false,
			Error:   err.Error(),
		}
	}

	// Add certificate info to the domain entry's metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	entry.Metadata["cert"] = certInfo

	return nil
}

// Close cleans up resources
func (p *Plugin) Close() error {
	return nil
}

// getCertInfo reads and validates the certificate for a domain
func (p *Plugin) getCertInfo(domain string) (*CertInfo, error) {
	certPath := filepath.Join(p.certDir, domain, "cert.pem")
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Validate certificate
	now := time.Now()
	isValid := now.After(cert.NotBefore) && now.Before(cert.NotAfter)

	return &CertInfo{
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		SerialNumber: cert.SerialNumber.String(),
		Issuer:       cert.Issuer.CommonName,
		IsValid:      isValid,
	}, nil
}

// New creates a new certs plugin instance
func New() *Plugin {
	return &Plugin{}
}
