package model

import "time"

// CertInfo represents certificate information for a domain
type CertInfo struct {
	IsValid   bool      // Whether the certificate is valid
	Issuer    string    // The certificate issuer
	Subject   string    // The certificate subject
	NotBefore time.Time // The certificate validity start time
	NotAfter  time.Time // The certificate validity end time
	Error     string    // Error message if the certificate is invalid
}
