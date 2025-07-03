package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Middleware creates a new authentication middleware
func Middleware(cfg *Config, logger *zap.Logger) fiber.Handler {
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
