package server

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

// CertInfo holds certificate information
type CertInfo struct {
	NotBefore    time.Time
	NotAfter     time.Time
	SerialNumber string
	Issuer       string
	IsValid      bool
	Error        string
}

// Server implements the gRPC plugin service
type Server struct {
	pb.UnimplementedPluginServer
	certDir string
}

// NewServer creates a new gRPC plugin server
func NewServer() *Server {
	return &Server{}
}

// Initialize implements the Initialize RPC
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	s.certDir = req.CertDir
	return &pb.InitializeResponse{
		Success: true,
	}, nil
}

// EnrichDomainEntry implements the EnrichDomainEntry RPC
func (s *Server) EnrichDomainEntry(ctx context.Context, req *pb.EnrichDomainEntryRequest) (*pb.EnrichDomainEntryResponse, error) {
	certInfo, err := s.getCertInfo(req.Entry.Domain)
	if err != nil {
		return &pb.EnrichDomainEntryResponse{
			Entry:   req.Entry,
			Success: true,
			Error:   err.Error(),
		}, nil
	}

	// Convert cert info to metadata
	metadata := make(map[string]string)
	metadata["cert.not_before"] = certInfo.NotBefore.Format(time.RFC3339)
	metadata["cert.not_after"] = certInfo.NotAfter.Format(time.RFC3339)
	metadata["cert.serial_number"] = certInfo.SerialNumber
	metadata["cert.issuer"] = certInfo.Issuer
	metadata["cert.is_valid"] = fmt.Sprintf("%v", certInfo.IsValid)
	if certInfo.Error != "" {
		metadata["cert.error"] = certInfo.Error
	}

	req.Entry.Metadata = metadata
	return &pb.EnrichDomainEntryResponse{
		Entry:   req.Entry,
		Success: true,
	}, nil
}

// Close implements the Close RPC
func (s *Server) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{
		Success: true,
	}, nil
}

// getCertInfo reads and validates the certificate for a domain
func (s *Server) getCertInfo(domain string) (*CertInfo, error) {
	certPath := filepath.Join(s.certDir, domain, "cert.pem")
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

// Serve starts the gRPC server
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPluginServer(grpcServer, s)

	return grpcServer.Serve(lis)
}
