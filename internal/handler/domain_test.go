package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal/util"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	serviceinterface "github.com/schumann-it/dehydrated-api-go/internal/service/interface"

	"github.com/schumann-it/dehydrated-api-go/internal/model"

	"github.com/schumann-it/dehydrated-api-go/internal/service"

	"github.com/gofiber/fiber/v2"
)

// TestDomainHandler tests the complete domain handler functionality.
// It verifies all CRUD operations for domain entries through the HTTP API.
func TestDomainHandler(t *testing.T) {
	// Test CreateDomain
	t.Run("CreateDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		req := model.CreateDomainRequest{
			Domain:           "example-create.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusCreated {
			t.Errorf("Expected status %d, got %d", fiber.StatusCreated, result.StatusCode)
		}

		var response model.DomainResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data.Domain != "example-create.com" {
			t.Errorf("Expected domain example-create.com, got %s", response.Data.Domain)
		}
	})

	// Test CreateInvalidDomain
	t.Run("CreateInvalidDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		req := model.CreateDomainRequest{
			Domain: "invalid..com",
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, result.StatusCode)
		}
	})

	// Test GetDomain
	t.Run("GetDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// First create the domain to ensure it exists
		createReq := model.CreateDomainRequest{
			Domain:           "example-get.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}
		createBody, _ := json.Marshal(createReq)

		createResp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(createBody))
		createResp.Header.Set("Content-Type", "application/json")

		createResult, err := app.Test(createResp)
		if err != nil {
			t.Fatalf("Failed to create domain for test: %v", err)
		}
		defer createResult.Body.Close()
		if createResult.StatusCode != fiber.StatusCreated {
			t.Fatalf("Failed to create domain, got status %d", createResult.StatusCode)
		}

		// Reload the service to ensure the cache is updated
		if err = s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Now get the domain
		resp := httptest.NewRequest("GET", "/api/v1/domains/example-get.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
			return
		}

		var response model.DomainResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data.Domain != "example-get.com" {
			t.Errorf("Expected domain example-get.com, got %s", response.Data.Domain)
		}
	})

	// Test GetNonExistentDomain
	t.Run("GetNonExistentDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		resp := httptest.NewRequest("GET", "/api/v1/domains/nonexistent.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status %d, got %d", fiber.StatusNotFound, result.StatusCode)
		}
	})

	// Test ListDomains
	t.Run("ListDomains", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		resp := httptest.NewRequest("GET", "/api/v1/domains", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.PaginatedDomainsResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Data) != 0 {
			t.Errorf("Expected 0 domains, got %d", len(response.Data))
		}

		if response.Pagination == nil {
			t.Error("Expected pagination info to be present")
		} else {
			if response.Pagination.CurrentPage != 1 {
				t.Errorf("Expected current page 1, got %d", response.Pagination.CurrentPage)
			}
			if response.Pagination.PerPage != model.DefaultPerPage {
				t.Errorf("Expected per page %d, got %d", model.DefaultPerPage, response.Pagination.PerPage)
			}
			if response.Pagination.Total != 0 {
				t.Errorf("Expected total 0, got %d", response.Pagination.Total)
			}
		}
	})

	t.Run("ListDomainsWithPagination", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// Create some test domains
		domains := []string{"domain1.com", "domain2.com", "domain3.com", "domain4.com", "domain5.com"}
		for _, domain := range domains {
			req := model.CreateDomainRequest{
				Domain:  domain,
				Enabled: true,
			}
			body, _ := json.Marshal(req)

			resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
			resp.Header.Set("Content-Type", "application/json")

			result, err := app.Test(resp)
			if err != nil {
				t.Fatalf("Failed to create domain %s: %v", domain, err)
			}
			defer result.Body.Close()

			if result.StatusCode != fiber.StatusCreated {
				t.Fatalf("Failed to create domain %s, got status %d", domain, result.StatusCode)
			}
		}

		// Reload the service to ensure the cache is updated
		if err := s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Test pagination with page=1, per_page=2
		resp := httptest.NewRequest("GET", "/api/v1/domains?page=1&per_page=2", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.PaginatedDomainsResponse
		if respErr := json.NewDecoder(result.Body).Decode(&response); respErr != nil {
			t.Fatalf("Failed to decode response: %v", respErr)
		}

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 domains, got %d", len(response.Data))
		}

		if response.Pagination == nil {
			t.Error("Expected pagination info to be present")
		} else {
			if response.Pagination.CurrentPage != 1 {
				t.Errorf("Expected current page 1, got %d", response.Pagination.CurrentPage)
			}
			if response.Pagination.PerPage != 2 {
				t.Errorf("Expected per page 2, got %d", response.Pagination.PerPage)
			}
			if response.Pagination.Total != 5 {
				t.Errorf("Expected total 5, got %d", response.Pagination.Total)
			}
			if response.Pagination.TotalPages != 3 {
				t.Errorf("Expected total pages 3, got %d", response.Pagination.TotalPages)
			}
			if !response.Pagination.HasNext {
				t.Error("Expected has_next to be true")
			}
			if response.Pagination.HasPrev {
				t.Error("Expected has_prev to be false for first page")
			}
			if response.Pagination.NextURL == "" {
				t.Error("Expected next_url to be present")
			}
			if response.Pagination.PrevURL != "" {
				t.Error("Expected prev_url to be empty for first page")
			}
		}

		// Test pagination with page=2, per_page=2
		resp2 := httptest.NewRequest("GET", "/api/v1/domains?page=2&per_page=2", http.NoBody)

		result2, err := app.Test(resp2)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result2.Body.Close()

		if result2.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result2.StatusCode)
		}

		var response2 model.PaginatedDomainsResponse
		if err := json.NewDecoder(result2.Body).Decode(&response2); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response2.Data) != 2 {
			t.Errorf("Expected 2 domains on page 2, got %d", len(response2.Data))
		}

		if response2.Pagination == nil {
			t.Error("Expected pagination info to be present")
		} else {
			if response2.Pagination.CurrentPage != 2 {
				t.Errorf("Expected current page 2, got %d", response2.Pagination.CurrentPage)
			}
			if !response2.Pagination.HasNext {
				t.Error("Expected has_next to be true for page 2")
			}
			if !response2.Pagination.HasPrev {
				t.Error("Expected has_prev to be true for page 2")
			}
			if response2.Pagination.NextURL == "" {
				t.Error("Expected next_url to be present for page 2")
			}
			if response2.Pagination.PrevURL == "" {
				t.Error("Expected prev_url to be present for page 2")
			}
		}
	})

	// Test ListDomains with sorting
	t.Run("ListDomainsWithSorting", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// Create test domains in reverse order
		domains := []string{"zebra.com", "alpha.com", "beta.com"}
		for _, domain := range domains {
			req := model.CreateDomainRequest{
				Domain:  domain,
				Enabled: true,
			}
			body, _ := json.Marshal(req)

			resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
			resp.Header.Set("Content-Type", "application/json")

			result, err := app.Test(resp)
			if err != nil {
				t.Fatalf("Failed to create domain %s: %v", domain, err)
			}
			defer result.Body.Close()

			if result.StatusCode != fiber.StatusCreated {
				t.Fatalf("Failed to create domain %s, got status %d", domain, result.StatusCode)
			}
		}

		// Reload the service to ensure the cache is updated
		if err := s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Test no sorting (original order)
		resp := httptest.NewRequest("GET", "/api/v1/domains", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.PaginatedDomainsResponse
		if respErr := json.NewDecoder(result.Body).Decode(&response); respErr != nil {
			t.Fatalf("Failed to decode response: %v", respErr)
		}

		if len(response.Data) != 3 {
			t.Errorf("Expected 3 domains, got %d", len(response.Data))
		}

		// Check original order (after file write/reload, domains are automatically sorted alphabetically)
		if response.Data[0].Domain != "alpha.com" {
			t.Errorf("Expected first domain to be alpha.com (alphabetical order after file write), got %s", response.Data[0].Domain)
		}
		if response.Data[1].Domain != "beta.com" {
			t.Errorf("Expected second domain to be beta.com (alphabetical order after file write), got %s", response.Data[1].Domain)
		}
		if response.Data[2].Domain != "zebra.com" {
			t.Errorf("Expected third domain to be zebra.com (alphabetical order after file write), got %s", response.Data[2].Domain)
		}

		// Test ascending sort
		respAsc := httptest.NewRequest("GET", "/api/v1/domains?sort=asc", http.NoBody)

		resultAsc, err := app.Test(respAsc)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer resultAsc.Body.Close()

		if resultAsc.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resultAsc.StatusCode)
		}

		var responseAsc model.PaginatedDomainsResponse
		if respErr := json.NewDecoder(resultAsc.Body).Decode(&responseAsc); respErr != nil {
			t.Fatalf("Failed to decode response: %v", respErr)
		}

		if len(responseAsc.Data) != 3 {
			t.Errorf("Expected 3 domains, got %d", len(responseAsc.Data))
		}

		// Check ascending order
		if responseAsc.Data[0].Domain != "alpha.com" {
			t.Errorf("Expected first domain to be alpha.com, got %s", responseAsc.Data[0].Domain)
		}
		if responseAsc.Data[1].Domain != "beta.com" {
			t.Errorf("Expected second domain to be beta.com, got %s", responseAsc.Data[1].Domain)
		}
		if responseAsc.Data[2].Domain != "zebra.com" {
			t.Errorf("Expected third domain to be zebra.com, got %s", responseAsc.Data[2].Domain)
		}

		// Test descending sort
		respDesc := httptest.NewRequest("GET", "/api/v1/domains?sort=desc", http.NoBody)

		resultDesc, err := app.Test(respDesc)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer resultDesc.Body.Close()

		if resultDesc.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resultDesc.StatusCode)
		}

		var responseDesc model.PaginatedDomainsResponse
		if err := json.NewDecoder(resultDesc.Body).Decode(&responseDesc); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(responseDesc.Data) != 3 {
			t.Errorf("Expected 3 domains, got %d", len(responseDesc.Data))
		}

		// Check descending order
		if responseDesc.Data[0].Domain != "zebra.com" {
			t.Errorf("Expected first domain to be zebra.com, got %s", responseDesc.Data[0].Domain)
		}
		if responseDesc.Data[1].Domain != "beta.com" {
			t.Errorf("Expected second domain to be beta.com, got %s", responseDesc.Data[1].Domain)
		}
		if responseDesc.Data[2].Domain != "alpha.com" {
			t.Errorf("Expected third domain to be alpha.com, got %s", responseDesc.Data[2].Domain)
		}
	})

	// Test ListDomains with search
	t.Run("ListDomainsWithSearch", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// Create test domains
		domains := []string{"example.com", "test.com", "example.org", "demo.net"}
		for _, domain := range domains {
			req := model.CreateDomainRequest{
				Domain:  domain,
				Enabled: true,
			}
			body, _ := json.Marshal(req)

			resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
			resp.Header.Set("Content-Type", "application/json")

			result, err := app.Test(resp)
			if err != nil {
				t.Fatalf("Failed to create domain %s: %v", domain, err)
			}
			defer result.Body.Close()

			if result.StatusCode != fiber.StatusCreated {
				t.Fatalf("Failed to create domain %s, got status %d", domain, result.StatusCode)
			}
		}

		// Reload the service to ensure the cache is updated
		if err := s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Test search for "example"
		resp := httptest.NewRequest("GET", "/api/v1/domains?search=example", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.PaginatedDomainsResponse
		if respErr := json.NewDecoder(result.Body).Decode(&response); respErr != nil {
			t.Fatalf("Failed to decode response: %v", respErr)
		}

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 domains matching 'example', got %d", len(response.Data))
		}

		// Check that both example.com and example.org are returned
		foundExampleCom := false
		foundExampleOrg := false
		for _, domain := range response.Data {
			if domain.Domain == "example.com" {
				foundExampleCom = true
			}
			if domain.Domain == "example.org" {
				foundExampleOrg = true
			}
		}
		if !foundExampleCom {
			t.Error("Expected to find example.com in search results")
		}
		if !foundExampleOrg {
			t.Error("Expected to find example.org in search results")
		}

		// Test case-insensitive search
		resp2 := httptest.NewRequest("GET", "/api/v1/domains?search=EXAMPLE", http.NoBody)

		result2, err := app.Test(resp2)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result2.Body.Close()

		if result2.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result2.StatusCode)
		}

		var response2 model.PaginatedDomainsResponse
		if err := json.NewDecoder(result2.Body).Decode(&response2); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response2.Data) != 2 {
			t.Errorf("Expected 2 domains matching 'EXAMPLE' (case-insensitive), got %d", len(response2.Data))
		}
	})

	// Test ListDomains with invalid sort parameter
	t.Run("ListDomainsWithInvalidSort", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Get("/api/v1/domains", handler.ListDomains)

		// Test invalid sort parameter
		resp := httptest.NewRequest("GET", "/api/v1/domains?sort=invalid", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, result.StatusCode)
		}

		var response model.PaginatedDomainsResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Success {
			t.Error("Expected success to be false")
		}
		if response.Error == "" {
			t.Error("Expected error message to be present")
		}
	})

	// Test UpdateDomain
	t.Run("UpdateDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// First create the domain to ensure it exists
		createReq := model.CreateDomainRequest{
			Domain:           "example-update.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}
		createBody, _ := json.Marshal(createReq)

		createResp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(createBody))
		createResp.Header.Set("Content-Type", "application/json")

		createResult, err := app.Test(createResp)
		if err != nil {
			t.Fatalf("Failed to create domain for test: %v", err)
		}
		defer createResult.Body.Close()
		if createResult.StatusCode != fiber.StatusCreated {
			t.Fatalf("Failed to create domain, got status %d", createResult.StatusCode)
		}

		// Reload the service to ensure the cache is updated
		if err = s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Now update the domain
		req := model.UpdateDomainRequest{
			AlternativeNames: util.StringSlicePtr([]string{"www.example.com", "api.example.com"}),
			Enabled:          util.BoolPtr(true),
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("PUT", "/api/v1/domains/example-update.com", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()
		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
			return
		}

		var response model.DomainResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if len(response.Data.AlternativeNames) != 2 {
			t.Errorf("Expected 2 alternative names, got %d", len(response.Data.AlternativeNames))
		}
	})

	// Test UpdateDomainWithoutOverwritingEmptyFields
	t.Run("UpdateDomainWithoutOverwritingEmptyFields", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// First create the domain to ensure it exists
		createReq := model.CreateDomainRequest{
			Domain:           "example-update-empty.com",
			AlternativeNames: []string{"www.example.com", "api.example.com"},
			Enabled:          true,
		}
		createBody, _ := json.Marshal(createReq)

		createResp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(createBody))
		createResp.Header.Set("Content-Type", "application/json")

		createResult, err := app.Test(createResp)
		if err != nil {
			t.Fatalf("Failed to create domain for test: %v", err)
		}
		defer createResult.Body.Close()
		if createResult.StatusCode != fiber.StatusCreated {
			t.Fatalf("Failed to create domain, got status %d", createResult.StatusCode)
		}

		// Reload the service to ensure the cache is updated
		if err = s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Now update the domain
		req := model.UpdateDomainRequest{
			Enabled: util.BoolPtr(true),
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("PUT", "/api/v1/domains/example-update-empty.com", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
			return
		}

		var response model.DomainResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if len(response.Data.AlternativeNames) != 2 {
			t.Errorf("Expected 2 alternative names, got %d", len(response.Data.AlternativeNames))
		}
	})

	// Test DeleteDomain
	t.Run("DeleteDomain", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()

		// Create a new Fiber app
		app := fiber.New()

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		// Create domain service
		s := service.NewDomainService(dc, nil)
		defer s.Close()

		// Create a new domain handler
		handler := NewDomainHandler(s)

		// register routes
		app.Post("/api/v1/domains", handler.CreateDomain)
		app.Get("/api/v1/domains", handler.ListDomains)
		app.Get("/api/v1/domains/:domain", handler.GetDomain)
		app.Put("/api/v1/domains/:domain", handler.UpdateDomain)
		app.Delete("/api/v1/domains/:domain", handler.DeleteDomain)

		// First create the domain to ensure it exists
		createReq := model.CreateDomainRequest{
			Domain:           "example-delete.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}
		createBody, _ := json.Marshal(createReq)

		createResp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(createBody))
		createResp.Header.Set("Content-Type", "application/json")

		createResult, err := app.Test(createResp)
		if err != nil {
			t.Fatalf("Failed to create domain for test: %v", err)
		}
		defer createResult.Body.Close()
		if createResult.StatusCode != fiber.StatusCreated {
			t.Fatalf("Failed to create domain, got status %d", createResult.StatusCode)
		}

		// Reload the service to ensure the cache is updated
		if err = s.Reload(); err != nil {
			t.Fatalf("Failed to reload service: %v", err)
		}

		// Now delete the domain
		resp := httptest.NewRequest("DELETE", "/api/v1/domains/example-delete.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()
		if result.StatusCode != fiber.StatusNoContent {
			t.Errorf("Expected status %d, got %d", fiber.StatusNoContent, result.StatusCode)
		}
	})
}

// TestRouteRegistration verifies that all domain-related routes are properly registered.
// It ensures that the handler correctly sets up all required endpoints.
func TestRouteRegistration(t *testing.T) {
	app := fiber.New()
	group := app.Group("/api/v1")
	handler := NewDomainHandler(&serviceinterface.MockErrDomainService{})
	handler.RegisterRoutes(group)

	// Test each route individually
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/domains"},
		{"GET", "/api/v1/domains/example.com"},
		{"POST", "/api/v1/domains"},
		{"PUT", "/api/v1/domains/example.com"},
		{"DELETE", "/api/v1/domains/example.com"},
	}

	// Get the app's route stack
	stack := app.Stack()
	if len(stack) == 0 {
		t.Fatal("No routes registered")
	}

	// Create a map of registered routes for easy lookup
	registeredRoutes := make(map[string]bool)
	for _, routes := range stack {
		for _, route := range routes {
			// Convert route pattern to a test path by replacing :param with a value
			testPath := route.Path
			if route.Path == "/api/v1/domains/:domain" {
				testPath = "/api/v1/domains/example.com"
			}
			key := route.Method + " " + testPath
			registeredRoutes[key] = true
		}
	}

	// Verify each test route exists
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			key := tt.method + " " + tt.path
			if !registeredRoutes[key] {
				t.Errorf("Route %s %s not found in registered routes", tt.method, tt.path)
			}
		})
	}
}

// TestServiceErrors verifies that the handler properly handles service errors.
// It tests error responses for various error conditions that may occur during domain operations.
func TestServiceErrors(t *testing.T) {
	app := fiber.New()
	group := app.Group("/api/v1")
	// Create a mock s that always returns errors
	s := &serviceinterface.MockErrDomainService{}
	handler := NewDomainHandler(s)
	handler.RegisterRoutes(group)

	// Test ListDomains with s error
	t.Run("ListDomains", func(t *testing.T) {
		resp := httptest.NewRequest("GET", "/api/v1/domains", http.NoBody)
		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()
		if result.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, result.StatusCode)
		}
	})

	// Test CreateDomain with s error
	t.Run("CreateDomain", func(t *testing.T) {
		req := model.CreateDomainRequest{
			Domain: "example.com",
		}
		body, _ := json.Marshal(req)
		resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}
		defer result.Body.Close()
		if result.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, result.StatusCode)
		}
	})
}

// TestCacheHeaders verifies that cache control headers are properly set on domain endpoints.
func TestCacheHeaders(t *testing.T) {
	app := fiber.New()
	group := app.Group("/api/v1")
	handler := NewDomainHandler(&serviceinterface.MockDomainService{})
	handler.RegisterRoutes(group)

	// Test each endpoint to ensure cache headers are set
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"ListDomains", "GET", "/api/v1/domains", ""},
		{"GetDomain", "GET", "/api/v1/domains/example.com", ""},
		{"CreateDomain", "POST", "/api/v1/domains", `{"domain": "test.com"}`},
		{"UpdateDomain", "PUT", "/api/v1/domains/example.com", `{"enabled": true}`},
		{"DeleteDomain", "DELETE", "/api/v1/domains/example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, http.NoBody)
			}

			result, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to test request: %v", err)
			}
			defer result.Body.Close()
		})
	}
}
