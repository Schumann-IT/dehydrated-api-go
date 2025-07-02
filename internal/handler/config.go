package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// ConfigHandler handles HTTP requests for Config operations
type ConfigHandler struct {
	cfg *dehydrated.Config
}

// NewConfigHandler creates a new ConfigHandler instance
func NewConfigHandler(cfg *dehydrated.Config) *ConfigHandler {
	return &ConfigHandler{
		cfg: cfg,
	}
}

// RegisterRoutes registers all Config-related routes
func (h *ConfigHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/config", h.Config)
}

// @Summary Get dehydrated configuration
// @Description Retrieve the current dehydrated configuration settings including paths, certificates, and operational parameters
// @Tags config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.ConfigResponse "Configuration retrieved successfully"
// @Failure 401 {object} model.ConfigResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 500 {object} model.ConfigResponse "Internal Server Error - Failed to retrieve configuration"
// @Router /config [get]
func (h *ConfigHandler) Config(c *fiber.Ctx) error {
	return c.JSON(model.ConfigResponse{
		Success: true,
		Data:    h.cfg,
	})
}
