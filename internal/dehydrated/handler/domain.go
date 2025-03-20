package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/service"
)

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	service *service.DomainService
}

// NewDomainHandler creates a new DomainHandler instance
func NewDomainHandler(service *service.DomainService) *DomainHandler {
	return &DomainHandler{
		service: service,
	}
}

// RegisterRoutes registers all domain-related routes
func (h *DomainHandler) RegisterRoutes(app *fiber.App) {
	domains := app.Group("/api/v1/domains")
	domains.Get("/", h.ListDomains)
	domains.Get("/:domain", h.GetDomain)
	domains.Post("/", h.CreateDomain)
	domains.Put("/:domain", h.UpdateDomain)
	domains.Delete("/:domain", h.DeleteDomain)
}

// ListDomains handles GET /api/v1/domains
func (h *DomainHandler) ListDomains(c *fiber.Ctx) error {
	entries, err := h.service.ListDomains()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.DomainsResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(model.DomainsResponse{
		Success: true,
		Data:    entries,
	})
}

// GetDomain handles GET /api/v1/domains/:domain
func (h *DomainHandler) GetDomain(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   "domain parameter is required",
		})
	}

	entry, err := h.service.GetDomain(domain)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(model.DomainResponse{
		Success: true,
		Data:    *entry,
	})
}

// CreateDomain handles POST /api/v1/domains
func (h *DomainHandler) CreateDomain(c *fiber.Ctx) error {
	var req model.CreateDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	entry, err := h.service.CreateDomain(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.DomainResponse{
		Success: true,
		Data:    *entry,
	})
}

// UpdateDomain handles PUT /api/v1/domains/:domain
func (h *DomainHandler) UpdateDomain(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   "domain parameter is required",
		})
	}

	var req model.UpdateDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	entry, err := h.service.UpdateDomain(domain, req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(model.DomainResponse{
		Success: true,
		Data:    *entry,
	})
}

// DeleteDomain handles DELETE /api/v1/domains/:domain
func (h *DomainHandler) DeleteDomain(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
			Success: false,
			Error:   "domain parameter is required",
		})
	}

	if err := h.service.DeleteDomain(domain); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
