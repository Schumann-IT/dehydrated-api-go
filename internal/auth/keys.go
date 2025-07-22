package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JWK represents a JSON Web Key containing information about a cryptographic key in JSON format.
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a set of JSON Web Keys, typically used to expose public keys for token verification.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// KeyManager handles fetching and caching of Azure AD public keys
type KeyManager struct {
	tenantID  string
	logger    *zap.Logger
	keys      map[string]*rsa.PublicKey
	mu        sync.RWMutex
	lastFetch time.Time
	cacheTTL  time.Duration
}

// NewKeyManager creates a new KeyManager instance
func NewKeyManager(tenantID string, logger *zap.Logger, cacheTTL string) *KeyManager {
	if tenantID == "" {
		logger.Error("tenantID must not be empty")
		return nil
	}

	// Parse the cache TTL
	ttl, err := time.ParseDuration(cacheTTL)
	if err != nil {
		logger.Warn("Invalid key cache TTL, using default 24h",
			zap.String("provided_ttl", cacheTTL),
			zap.Error(err),
		)
		ttl = 24 * time.Hour
	}

	return &KeyManager{
		tenantID: tenantID,
		logger:   logger,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: ttl,
	}
}

// GetKey retrieves a public key by its key ID
func (km *KeyManager) GetKey(kid string) (*rsa.PublicKey, error) {
	km.mu.RLock()
	key, exists := km.keys[kid]
	km.mu.RUnlock()

	if exists {
		return key, nil
	}

	// Key not found, refresh the cache
	if err := km.refreshKeys(); err != nil {
		return nil, fmt.Errorf("failed to refresh keys: %w", err)
	}

	km.mu.RLock()
	key, exists = km.keys[kid]
	km.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

// refreshKeys fetches the latest public keys from Azure AD
func (km *KeyManager) refreshKeys() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Check if we need to refresh (cache TTL)
	if time.Since(km.lastFetch) < km.cacheTTL && len(km.keys) > 0 {
		km.logger.Debug("Keys cache is still valid, skipping refresh")
		return nil
	}

	km.logger.Info("Fetching Azure AD public keys",
		zap.String("tenant_id", km.tenantID),
	)

	// Fetch keys from Azure AD
	jwksURL := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", km.tenantID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch keys, status: %d", resp.StatusCode)
	}

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Clear existing keys
	km.keys = make(map[string]*rsa.PublicKey)

	// Parse and store keys
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			km.logger.Debug("Skipping non-RSA key",
				zap.String("kid", jwk.Kid),
				zap.String("kty", jwk.Kty),
			)
			continue
		}

		publicKey, err := km.parseRSAPublicKey(&jwk)
		if err != nil {
			km.logger.Warn("Failed to parse RSA public key",
				zap.String("kid", jwk.Kid),
				zap.Error(err),
			)
			continue
		}

		km.keys[jwk.Kid] = publicKey
		km.logger.Debug("Added public key",
			zap.String("kid", jwk.Kid),
			zap.String("use", jwk.Use),
			zap.String("alg", jwk.Alg),
		)
	}

	km.lastFetch = time.Now()
	km.logger.Info("Successfully refreshed public keys",
		zap.Int("key_count", len(km.keys)),
		zap.Time("last_fetch", km.lastFetch),
	)

	return nil
}

// parseRSAPublicKey converts a JWK to an RSA public key
func (km *KeyManager) parseRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (N)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent (E)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}
