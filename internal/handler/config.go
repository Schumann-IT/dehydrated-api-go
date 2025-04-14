package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

// ConfigHandler handles HTTP requests for dehydrated configuration operations.
// It provides endpoints for retrieving and managing the dehydrated configuration.
type ConfigHandler struct {
	service *service.DomainService
}

// NewConfigHandler creates a new ConfigHandler instance with the given domain service.
// The domain service is used to access the dehydrated configuration.
func NewConfigHandler(service *service.DomainService) *ConfigHandler {
	return &ConfigHandler{
		service: service,
	}
}

// RegisterRoutes registers all config-related routes with the given Fiber application.
// It sets up the routes under the /api/v1/config prefix.
func (h *ConfigHandler) RegisterRoutes(app *fiber.App) {
	g := app.Group("/api/v1/config")
	g.Get("/dehydrated", h.GetDehydratedConfig)
}

// GetDehydratedConfig handles GET /api/v1/config/dehydrated requests.
// It returns the current dehydrated configuration if available, or a 404 error if not.
func (h *ConfigHandler) GetDehydratedConfig(c *fiber.Ctx) error {
	if h.service.Registry != nil {
		return c.Status(fiber.StatusOK).JSON(h.service.Registry.Config)
	}

	return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
		Success: false,
		Error:   "Could not load config",
	})
}
