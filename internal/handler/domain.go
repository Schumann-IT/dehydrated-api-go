package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// DomainService defines the interface for domain operations
type DomainService interface {
	ListDomains() ([]model.DomainEntry, error)
	GetDomain(domain string) (*model.DomainEntry, error)
	CreateDomain(req model.CreateDomainRequest) (*model.DomainEntry, error)
	UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error)
	DeleteDomain(domain string) error
	Close() error
}

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	service DomainService
}

// NewDomainHandler creates a new DomainHandler instance
func NewDomainHandler(service DomainService) *DomainHandler {
	return &DomainHandler{
		service: service,
	}
}

// RegisterRoutes registers all domain-related routes
func (h *DomainHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/domains", h.ListDomains)
	app.Get("/api/v1/domains/:domain", h.GetDomain)
	app.Post("/api/v1/domains", h.CreateDomain)
	app.Put("/api/v1/domains/:domain", h.UpdateDomain)
	app.Delete("/api/v1/domains/:domain", h.DeleteDomain)
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
