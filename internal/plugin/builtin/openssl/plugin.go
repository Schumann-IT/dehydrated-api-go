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

// Plugin implements the openssl metadata plugin
type Plugin struct {
	pb.UnimplementedPluginServer
	dehydratedConfig *pb.DehydratedConfig
}

// CertificateInfo represents the information about a certificate
type CertificateInfo struct {
	File      string    `json:"file"`
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	KeyType   string    `json:"key_type"`
	KeySize   int       `json:"key_size"`
}

// New creates a new openssl plugin instance
func New() *Plugin {
	return &Plugin{}
}

// Initialize initializes the plugin with configuration
func (p *Plugin) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	p.dehydratedConfig = req.DehydratedConfig
	return &pb.InitializeResponse{}, nil
}

// GetMetadata returns metadata for the given domain
func (p *Plugin) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	metadata := make(map[string]*structpb.Value)

	// Get the domain's certificate directory
	domainDir := filepath.Join(p.dehydratedConfig.CertDir, req.Domain)
	if _, err := os.Stat(domainDir); os.IsNotExist(err) {
		return &pb.GetMetadataResponse{Metadata: metadata}, nil
	}

	// Find all certificate files
	certFiles, err := filepath.Glob(filepath.Join(domainDir, "*.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to find certificate files: %w", err)
	}

	certificates := make([]CertificateInfo, 0)
	for _, certFile := range certFiles {
		info, err := p.analyzeCertificate(certFile)
		if err != nil {
			continue // Skip invalid certificates
		}
		certificates = append(certificates, info)
	}

	// Convert certificates to metadata
	if len(certificates) > 0 {
		for _, certificate := range certificates {
			b, err := json.Marshal(certificate)
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
			metadata[certificate.File] = value
		}
	}

	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (p *Plugin) analyzeCertificate(certFile string) (CertificateInfo, error) {
	b, err := os.ReadFile(certFile)
	if err != nil {
		return CertificateInfo{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Decode the PEM block
	pb, _ := pem.Decode(b)
	if pb == nil {
		return CertificateInfo{}, fmt.Errorf("failed to decode PEM block")
	}
	cert, err := x509.ParseCertificate(pb.Bytes)
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

// Close cleans up any resources
func (p *Plugin) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}
