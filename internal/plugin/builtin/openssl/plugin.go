// Package openssl provides a built-in plugin for analyzing SSL/TLS certificates.
// It implements the Plugin interface to extract and provide metadata about certificates,
// including their validity periods, key types, and sizes.
package openssl

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin implements the openssl metadata plugin for analyzing SSL/TLS certificates.
// It can analyze certificate files (cert.pem), chain files (chain.pem), and full chain files (fullchain.pem)
// based on configuration settings.
type Plugin struct {
	pb.UnimplementedPluginServer
	dehydratedConfig *pb.DehydratedConfig
	cert             bool // whether to analyze cert.pem
	chain            bool // whether to analyze chain.pem
	fullchain        bool // whether to analyze fullchain.pem
}

// CertificateInfo represents the information extracted from a certificate file.
// All fields are exported and tagged for JSON serialization to support metadata exchange.
type CertificateInfo struct {
	File      string    `json:"file"`       // Path to the certificate file
	Subject   string    `json:"subject"`    // Certificate subject DN
	Issuer    string    `json:"issuer"`     // Certificate issuer DN
	NotBefore time.Time `json:"not_before"` // Start of validity period
	NotAfter  time.Time `json:"not_after"`  // End of validity period
	KeyType   string    `json:"key_type"`   // Type of public key (RSA/EC)
	KeySize   int       `json:"key_size"`   // Size of the public key in bits
}

// New creates a new openssl plugin instance with default settings.
// By default, only cert.pem analysis is enabled.
func New() *Plugin {
	return &Plugin{
		cert:      true,
		chain:     false,
		fullchain: false,
	}
}

// Initialize configures the plugin with the provided settings.
// It accepts boolean flags to control which certificate files to analyze:
// - cert: analyze cert.pem files
// - chain: analyze chain.pem files
// - fullchain: analyze fullchain.pem files
func (p *Plugin) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	if cert, ok := req.Config["cert"]; ok {
		if v, ok := cert.GetKind().(*structpb.Value_BoolValue); ok {
			p.cert = v.BoolValue
		}
	}
	if chain, ok := req.Config["chain"]; ok {
		if v, ok := chain.GetKind().(*structpb.Value_BoolValue); ok {
			p.chain = v.BoolValue
		}
	}
	if fullchain, ok := req.Config["fullchain"]; ok {
		if v, ok := fullchain.GetKind().(*structpb.Value_BoolValue); ok {
			p.fullchain = v.BoolValue
		}
	}
	p.dehydratedConfig = req.DehydratedConfig
	return &pb.InitializeResponse{}, nil
}

// GetMetadata analyzes certificate files for the specified domain and returns their metadata.
// It looks for certificate files in the configured certificate directory under the domain's subdirectory.
// The metadata includes certificate information such as validity period, key type, and size.
func (p *Plugin) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	metadata := make(map[string]*structpb.Value)

	// Get the domain's certificate directory
	domainDir := filepath.Join(p.dehydratedConfig.CertDir, req.Domain)
	if _, err := os.Stat(domainDir); os.IsNotExist(err) {
		return &pb.GetMetadataResponse{Metadata: metadata}, nil
	}

	certFiles := map[string]string{}
	if p.cert {
		certFiles["cert"] = filepath.Join(domainDir, "cert.pem")
	}
	if p.chain {
		certFiles["chain"] = filepath.Join(domainDir, "chain.pem")
	}
	if p.fullchain {
		certFiles["fullchain"] = filepath.Join(domainDir, "fullchain.pem")
	}

	for name, certFile := range certFiles {
		info, err := p.analyzeCertificate(certFile)
		if err != nil {
			continue // Skip invalid certificates
		}
		b, err := json.Marshal(info)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal certificate: %w", err)
		}
		certMap := make(map[string]interface{})
		err = json.Unmarshal(b, &certMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal certificate: %w", err)
		}
		value, err := structpb.NewValue(certMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal certificate: %w", err)
		}
		metadata[name] = value
	}

	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

// analyzeCertificate reads and parses a certificate file to extract its information.
// It supports both RSA and EC certificates, extracting details like subject, issuer,
// validity period, and key characteristics.
func (p *Plugin) analyzeCertificate(certFile string) (CertificateInfo, error) {
	b, err := os.ReadFile(certFile)
	if err != nil {
		return CertificateInfo{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Decode the PEM block
	bp, _ := pem.Decode(b)
	if bp == nil {
		return CertificateInfo{}, fmt.Errorf("failed to decode PEM block")
	}
	cert, err := x509.ParseCertificate(bp.Bytes)
	if err != nil {
		return CertificateInfo{}, err
	}

	keySize, keyType := getKeySize(cert.PublicKey)

	return CertificateInfo{
		File:      certFile,
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		KeyType:   keyType,
		KeySize:   keySize,
	}, nil
}

// getKeySize determines the key type and size from a public key.
// It supports RSA and ECDSA keys, returning the key size in bits and a string
// identifier for the key type ("rsaEncryption" or "ecPublicKey").
func getKeySize(pub interface{}) (int, string) {
	switch k := pub.(type) {
	case *rsa.PublicKey:
		return k.N.BitLen(), "rsaEncryption"
	case *ecdsa.PublicKey:
		// For ECDSA, we return the curve size
		return k.Curve.Params().BitSize, "ecPublicKey"
	default:
		return 0, "unknown"
	}
}

// Close performs cleanup when the plugin is being shut down.
// Currently, this is a no-op as the plugin doesn't maintain any resources that need cleanup.
func (p *Plugin) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}
