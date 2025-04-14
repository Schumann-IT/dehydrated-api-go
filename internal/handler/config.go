package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

// ConfigHandler handles HTTP requests for config operations
type ConfigHandler struct {
	service *service.DomainService
}

// NewConfigHandler creates a new ConfigHandler instance
func NewConfigHandler(service *service.DomainService) *ConfigHandler {
	return &ConfigHandler{
		service: service,
	}
}

// RegisterRoutes registers all config-related routes
func (h *ConfigHandler) RegisterRoutes(app *fiber.App) {
	g := app.Group("/api/v1/config")
	g.Get("/dehydrated", h.GetDehydratedConfig)
}

// GetDehydratedConfig handles GET /api/v1/config/dehydrated
func (h *ConfigHandler) GetDehydratedConfig(c *fiber.Ctx) error {
	if h.service.Registry != nil {
		return c.Status(fiber.StatusOK).JSON(h.service.Registry.Config)
	}

	return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
		Success: false,
		Error:   "Could not load config",
	})
}
