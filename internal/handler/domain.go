// Package handler provides HTTP handlers for the dehydrated-api-go application.
// It includes handlers for domain management and configuration operations.
package handler

import (
	"fmt"
	"strings"

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
// @Description Get a paginated list of all configured domains with optional sorting and searching
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (1-based, defaults to 1)" minimum(1)
// @Param per_page query int false "Number of items per page (defaults to 100, max 1000)" minimum(1) maximum(1000)
// @Param sort query string false "Sort order for domain field (asc or desc, optional - defaults to alphabetical order)" Enums(asc, desc)
// @Param search query string false "Search term to filter domains by domain field (case-insensitive contains)"
// @Success 200 {object} model.PaginatedDomainsResponse
// @Failure 400 {object} model.PaginatedDomainsResponse "Bad Request - Invalid pagination parameters"
// @Failure 401 {object} model.PaginatedDomainsResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 500 {object} model.PaginatedDomainsResponse "Internal Server Error"
// @Router /api/v1/domains [get]
// ListDomains handles GET /api/v1/domains
func (h *DomainHandler) ListDomains(c *fiber.Ctx) error {
	// Parse and validate pagination parameters
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", model.DefaultPerPage)

	// Parse sort and search parameters
	sortOrder := c.Query("sort", "")
	search := c.Query("search", "")

	// Validate page parameter
	if page < model.MinPage {
		return c.Status(fiber.StatusBadRequest).JSON(model.PaginatedDomainsResponse{
			Success: false,
			Error:   "page parameter must be at least 1",
		})
	}

	// Validate and cap per_page parameter
	if perPage < model.MinPerPage {
		perPage = model.MinPerPage
	} else if perPage > model.MaxPerPage {
		perPage = model.MaxPerPage
	}

	// Validate sort parameter (only if provided)
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return c.Status(fiber.StatusBadRequest).JSON(model.PaginatedDomainsResponse{
			Success: false,
			Error:   "sort parameter must be either 'asc' or 'desc'",
		})
	}

	// Get paginated domains from service
	entries, pagination, err := h.service.ListDomains(page, perPage, sortOrder, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.PaginatedDomainsResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Generate pagination URLs
	if pagination != nil {
		h.generatePaginationURLs(c, pagination)
	}

	return c.JSON(model.PaginatedDomainsResponse{
		Success:    true,
		Data:       entries,
		Pagination: pagination,
	})
}

// generatePaginationURLs generates the next and previous URLs for pagination
func (h *DomainHandler) generatePaginationURLs(c *fiber.Ctx, pagination *model.PaginationInfo) {
	baseURL := c.BaseURL() + c.Path()

	// Build query parameters
	queryParams := make(map[string]string)

	// Add existing query parameters (except pagination ones)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		if keyStr != "page" && keyStr != "per_page" {
			queryParams[keyStr] = string(value)
		}
	})

	// Always include per_page in URLs
	queryParams["per_page"] = fmt.Sprintf("%d", pagination.PerPage)

	// Generate next URL
	if pagination.HasNext {
		nextParams := make(map[string]string)
		for k, v := range queryParams {
			nextParams[k] = v
		}
		nextParams["page"] = fmt.Sprintf("%d", pagination.CurrentPage+1)
		pagination.NextURL = h.buildURL(baseURL, nextParams)
	}

	// Generate previous URL
	if pagination.HasPrev {
		prevParams := make(map[string]string)
		for k, v := range queryParams {
			prevParams[k] = v
		}
		prevParams["page"] = fmt.Sprintf("%d", pagination.CurrentPage-1)
		pagination.PrevURL = h.buildURL(baseURL, prevParams)
	}
}

// buildURL constructs a URL with query parameters
func (h *DomainHandler) buildURL(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}

	return baseURL + "?" + strings.Join(queryParts, "&")
}

// @Summary Get a domain
// @Description Get details of a specific domain
// @Tags domains
// @Produce json
// @Security BearerAuth
// @Param domain path string true "Domain name"
// @Param alias query string false "Optional alias to uniquely identify the domain entry"
// @Success 200 {object} model.DomainResponse
// @Failure 400 {object} model.DomainResponse "Bad Request - Invalid domain parameter"
// @Failure 401 {object} model.DomainResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 404 {object} model.DomainResponse "Not Found - Domain not found"
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

	entry, err := h.service.GetDomain(domain, c.Query("alias"))

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
// @Failure 400 {object} model.DomainResponse "Bad Request - Invalid request body or domain already exists"
// @Failure 401 {object} model.DomainResponse "Unauthorized - Invalid or missing authentication token"
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

	entry, err := h.service.CreateDomain(&req)
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
// @Failure 400 {object} model.DomainResponse "Bad Request - Invalid request body or domain parameter"
// @Failure 401 {object} model.DomainResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 404 {object} model.DomainResponse "Not Found - Domain not found"
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

	var entry *model.DomainEntry
	var err error

	entry, err = h.service.UpdateDomain(domain, req)
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
// @Param request body model.DeleteDomainRequest true "Domain delete request"
// @Success 204 "No Content"
// @Failure 400 {object} model.DomainResponse "Bad Request - Invalid domain parameter"
// @Failure 401 {object} model.DomainResponse "Unauthorized - Invalid or missing authentication token"
// @Failure 404 {object} model.DomainResponse "Not Found - Domain not found"
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

	var req model.DeleteDomainRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.DomainResponse{
				Success: false,
				Error:   "invalid request body",
			})
		}
	} else {
		// If no body is provided, use an empty DeleteDomainRequest
		req = model.DeleteDomainRequest{}
	}

	err := h.service.DeleteDomain(domain, req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.DomainResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
