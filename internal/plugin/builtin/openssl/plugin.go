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
	cert             bool
	chain            bool
	fullchain        bool
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
	return &Plugin{
		cert:      true,
		chain:     false,
		fullchain: false,
	}
}

// Initialize initializes the plugin with configuration
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

// GetMetadata returns metadata for the given domain
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
