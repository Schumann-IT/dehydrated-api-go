// Package handler provides HTTP handlers for the dehydrated-api-go application.
// It includes handlers for domain management and configuration operations.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	serviceinterface "github.com/schumann-it/dehydrated-api-go/internal/service/interface"
)

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	service serviceinterface.DomainService
}

// NewDomainHandler creates a new DomainHandler instance
func NewDomainHandler(service serviceinterface.DomainService) *DomainHandler {
	return &DomainHandler{
		service: service,
	}
}

// RegisterRoutes registers all domain-related routes
func (h *DomainHandler) RegisterRoutes(app fiber.Router) {
	app.Get("domains", h.ListDomains)
	app.Get("domains/:domain", h.GetDomain)
	app.Post("domains", h.CreateDomain)
	app.Put("domains/:domain", h.UpdateDomain)
	app.Delete("domains/:domain", h.DeleteDomain)
}

// @Summary List all domains
// @Description Get a list of all configured domains
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.DomainsResponse
// @Failure 500 {object} model.DomainsResponse
// @Router /api/v1/domains [get]
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

// @Summary Get a domain
// @Description Get details of a specific domain
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain path string true "Domain name"
// @Success 200 {object} model.DomainResponse
// @Failure 400 {object} model.DomainResponse
// @Failure 404 {object} model.DomainResponse
// @Router /api/v1/domains/{domain} [get]
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
		Data:    entry,
	})
}

// @Summary Create a domain
// @Description Create a new domain entry
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateDomainRequest true "Domain creation request"
// @Success 201 {object} model.DomainResponse
// @Failure 400 {object} model.DomainResponse
// @Router /api/v1/domains [post]
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
		Data:    entry,
	})
}

// @Summary Update a domain
// @Description Update an existing domain entry
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain path string true "Domain name"
// @Param request body model.UpdateDomainRequest true "Domain update request"
// @Success 200 {object} model.DomainResponse
// @Failure 400 {object} model.DomainResponse
// @Failure 404 {object} model.DomainResponse
// @Router /api/v1/domains/{domain} [put]
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
		Data:    entry,
	})
}

// @Summary Delete a domain
// @Description Delete a domain entry
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain path string true "Domain name"
// @Success 204 "No Content"
// @Failure 400 {object} model.DomainResponse
// @Failure 404 {object} model.DomainResponse
// @Router /api/v1/domains/{domain} [delete]
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
