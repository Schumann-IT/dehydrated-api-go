package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// HealthHandler handles HTTP requests for health operations
type HealthHandler struct {
	status bool
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{status: true}
}

// RegisterRoutes registers all health-related routes
func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.Health)
}

// @Summary Health check
// @Description Check if the API is running and healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.DomainsResponse
// @Router /health [get]
// Health handles GET /health
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(model.DomainsResponse{
		Success: h.status,
	})
}
