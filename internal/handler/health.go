package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// DomainHandler handles HTTP requests for helath operations
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// RegisterRoutes registers all health-related routes
func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/health", h.Health)
}

// Health handles GET /api/v1/health
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(model.DomainsResponse{
		Success: true,
	})
}
