package openssl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	File        string    `json:"file"`
	Subject     string    `json:"subject"`
	Issuer      string    `json:"issuer"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	KeyType     string    `json:"key_type"`
	KeySize     int       `json:"key_size"`
	Serial      string    `json:"serial"`
	Extensions  []string  `json:"extensions"`
	OCSPEnabled bool      `json:"ocsp_enabled"`
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

// analyzeCertificate analyzes a certificate file using OpenSSL
func (p *Plugin) analyzeCertificate(certFile string) (CertificateInfo, error) {
	info := CertificateInfo{}

	info.File = certFile

	// Read certificate information using OpenSSL
	cmd := exec.Command("openssl", "x509", "-in", certFile, "-noout", "-text")
	output, err := cmd.Output()
	if err != nil {
		return info, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Subject:") {
			info.Subject = strings.TrimSpace(strings.TrimPrefix(line, "Subject:"))
		} else if strings.HasPrefix(line, "Issuer:") {
			info.Issuer = strings.TrimSpace(strings.TrimPrefix(line, "Issuer:"))
		} else if strings.HasPrefix(line, "Not Before:") {
			info.NotBefore, err = time.Parse("Jan 2 15:04:05 2006 GMT", strings.TrimSpace(strings.TrimPrefix(line, "Not Before:")))
			if err != nil {
				return info, fmt.Errorf("failed to parse Not Before date: %w", err)
			}
		} else if strings.HasPrefix(line, "Not After :") {
			info.NotAfter, err = time.Parse("Jan 2 15:04:05 2006 GMT", strings.TrimSpace(strings.TrimPrefix(line, "Not After :")))
			if err != nil {
				return info, fmt.Errorf("failed to parse Not After date: %w", err)
			}
		} else if strings.HasPrefix(line, "Public Key Algorithm:") {
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "Public Key Algorithm:")), " ")
			if len(parts) >= 2 {
				info.KeyType = parts[0]
				if len(parts) > 2 {
					info.KeySize = parseKeySize(parts[2])
				}
			}
		} else if strings.HasPrefix(line, "Serial Number:") {
			info.Serial = strings.TrimSpace(strings.TrimPrefix(line, "Serial Number:"))
		} else if strings.HasPrefix(line, "X509v3 extensions:") {
			// Collect extensions
			for j := i + 1; j < len(lines); j++ {
				extLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(extLine, "---") {
					break
				}
				if strings.HasPrefix(extLine, "X509v3") {
					info.Extensions = append(info.Extensions, extLine)
					if strings.Contains(extLine, "OCSP Must Staple") {
						info.OCSPEnabled = true
					}
				}
			}
		}
	}

	return info, nil
}

// parseKeySize extracts the key size from a string like "RSA Public-Key: (2048 bit)"
func parseKeySize(s string) int {
	var size int
	fmt.Sscanf(s, "(%d bit)", &size)
	return size
}

// Close cleans up any resources
func (p *Plugin) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}
