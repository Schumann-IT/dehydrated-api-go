package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	// Test CreateDomain
	t.Run("CreateDomain", func(t *testing.T) {
		req := model.CreateDomainRequest{
			Domain:           "example.com",
			AlternativeNames: []string{"www.example.com"},
			Alias:            "*.example.com",
			Enabled:          true,
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("POST", "/api/v1/domains", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

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
		if response.Data.Domain != "example.com" {
			t.Errorf("Expected domain example.com, got %s", response.Data.Domain)
		}
	})

	// Test CreateInvalidDomain
	t.Run("CreateInvalidDomain", func(t *testing.T) {
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

		if result.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, result.StatusCode)
		}
	})

	// Test GetDomain
	t.Run("GetDomain", func(t *testing.T) {
		resp := httptest.NewRequest("GET", "/api/v1/domains/example.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.DomainResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data.Domain != "example.com" {
			t.Errorf("Expected domain example.com, got %s", response.Data.Domain)
		}
	})

	// Test GetNonExistentDomain
	t.Run("GetNonExistentDomain", func(t *testing.T) {
		resp := httptest.NewRequest("GET", "/api/v1/domains/nonexistent.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if result.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status %d, got %d", fiber.StatusNotFound, result.StatusCode)
		}
	})

	// Test ListDomains
	t.Run("ListDomains", func(t *testing.T) {
		resp := httptest.NewRequest("GET", "/api/v1/domains", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
		}

		var response model.DomainsResponse
		if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(response.Data))
		}
	})

	// Test UpdateDomain
	t.Run("UpdateDomain", func(t *testing.T) {
		req := model.UpdateDomainRequest{
			AlternativeNames: util.StringSlicePtr([]string{"www.example.com", "api.example.com"}),
			Enabled:          util.BoolPtr(true),
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("PUT", "/api/v1/domains/example.com", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if result.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, result.StatusCode)
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

	// Test UpdateDomain
	t.Run("UpdateDomainWithoutOverwritingEmptyFields", func(t *testing.T) {
		req := model.UpdateDomainRequest{
			Enabled: util.BoolPtr(true),
		}
		body, _ := json.Marshal(req)

		resp := httptest.NewRequest("PUT", "/api/v1/domains/example.com", bytes.NewReader(body))
		resp.Header.Set("Content-Type", "application/json")

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

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
		resp := httptest.NewRequest("DELETE", "/api/v1/domains/example.com", http.NoBody)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

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
		if result.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, result.StatusCode)
		}
	})
}
