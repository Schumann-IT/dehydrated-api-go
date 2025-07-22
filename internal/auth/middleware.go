package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Middleware creates new authentication middleware
func Middleware(cfg *Config, logger *zap.Logger) fiber.Handler {
	// Initialize the key manager if signature validation is enabled
	var keyManager *KeyManager
	if cfg.EnableSignatureValidation {
		keyManager = NewKeyManager(cfg.TenantID, logger, cfg.KeyCacheTTL)
		logger.Info("JWT signature validation enabled",
			zap.String("tenant_id", cfg.TenantID),
			zap.String("key_cache_ttl", cfg.KeyCacheTTL),
		)
	} else {
		logger.Warn("JWT signature validation disabled - using claim-based validation only")
	}

	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header format")
		}

		token, claims, _, err2, done := parseToken(parts, logger)
		if done {
			return err2
		}

		// Validate signature if enabled
		if cfg.EnableSignatureValidation && keyManager != nil {
			err, done1 := validateSignature(parts[1], keyManager, logger)
			if done1 {
				return err
			}
		}

		err, done2 := validateExpiration(claims, logger)
		if done2 {
			return err
		}

		err3, done3 := validateAudience(claims, logger, cfg)
		if done3 {
			return err3
		}

		err4, done4 := validateIssuer(claims, logger, cfg)
		if done4 {
			return err4
		}

		// Store the validated token in the context for later use
		c.Locals("token", token)

		return c.Next()
	}
}

func validateSignature(tokenString string, keyManager *KeyManager, logger *zap.Logger) (error, bool) {
	// Parse the token to get the header and extract the key ID
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		logger.Error("failed to parse token for signature validation",
			zap.Error(err),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token format"), true
	}

	// Extract key ID from the header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		logger.Error("missing key ID in token header")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token header"), true
	}

	// Get the public key for this key ID
	publicKey, err := keyManager.GetKey(kid)
	if err != nil {
		logger.Error("failed to get public key for signature validation",
			zap.String("kid", kid),
			zap.Error(err),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token signature"), true
	}

	// Validate the token signature
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return publicKey, nil
	})

	if err != nil {
		logger.Error("token signature validation failed",
			zap.String("kid", kid),
			zap.Error(err),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token signature"), true
	}

	if !parsedToken.Valid {
		logger.Error("token is invalid after signature validation",
			zap.String("kid", kid),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token"), true
	}

	logger.Debug("token signature validation successful",
		zap.String("kid", kid),
	)

	return nil, false
}

func validateIssuer(claims jwt.MapClaims, logger *zap.Logger, cfg *Config) (error, bool) {
	// Validate the issuer
	iss, ok := claims["iss"].(string)
	if !ok {
		logger.Error("missing issuer claim")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims"), true
	}

	// The issuer should be in the format: https://sts.windows.net/{tenantId}/
	if !strings.HasPrefix(iss, "https://sts.windows.net/") || !strings.HasSuffix(iss, "/") {
		logger.Error("invalid issuer format",
			zap.String("issuer", iss),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token issuer format"), true
	}

	// Extract tenant ID from issuer
	issuerParts := strings.Split(iss, "/")
	if len(issuerParts) < 4 {
		logger.Error("invalid issuer format",
			zap.String("issuer", iss),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token issuer format"), true
	}

	issuerTenantID := issuerParts[3]
	logger.Debug("comparing tenant IDs",
		zap.String("issuer_tenant", issuerTenantID),
		zap.String("configured_tenant", cfg.TenantID),
		zap.String("full_issuer", iss),
	)
	if issuerTenantID != cfg.TenantID {
		logger.Error("invalid issuer tenant",
			zap.String("issuer_tenant", issuerTenantID),
			zap.String("expected_tenant", cfg.TenantID),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token issuer tenant"), true
	}
	return nil, false
}

func validateAudience(claims jwt.MapClaims, logger *zap.Logger, cfg *Config) (error, bool) {
	// Validate the audience
	aud, ok := claims["aud"].(string)
	if !ok {
		logger.Error("missing audience claim")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims"), true
	}

	audienceAllowed := false
	for _, allowed := range cfg.AllowedAudiences {
		if aud == allowed {
			audienceAllowed = true
			break
		}
	}

	if !audienceAllowed {
		logger.Error("invalid audience",
			zap.String("audience", aud),
			zap.Strings("allowed_audiences", cfg.AllowedAudiences),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token audience"), true
	}
	return nil, false
}

func validateExpiration(claims jwt.MapClaims, logger *zap.Logger) (error, bool) {
	// Validate expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		logger.Error("missing expiration claim")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims"), true
	}

	expirationTime := time.Unix(int64(exp), 0)
	if time.Now().After(expirationTime) {
		logger.Error("token has expired",
			zap.Time("expiration_time", expirationTime),
			zap.Time("current_time", time.Now()),
		)
		return fiber.NewError(fiber.StatusUnauthorized, "token has expired"), true
	}
	return nil, false
}

func parseToken(parts []string, logger *zap.Logger) (string, jwt.MapClaims, bool, error, bool) {
	token := parts[1]

	// Parse the token without validating the signature first
	parser := jwt.Parser{}
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logger.Error("failed to parse token",
			zap.Error(err),
			zap.String("token", token),
		)
		return "", nil, false, fiber.NewError(fiber.StatusUnauthorized, "invalid token format"), true
	}

	// Get the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("invalid token claims")
		return "", nil, false, fiber.NewError(fiber.StatusUnauthorized, "invalid token claims"), true
	}
	return token, claims, ok, nil, false
}
