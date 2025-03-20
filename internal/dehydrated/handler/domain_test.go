package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/service"
)

func TestDomainHandler(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Create a new Fiber app
	app := fiber.New()

	// Create domain service
	service, err := service.NewDomainService(service.DomainServiceConfig{
		DomainsFile:   domainsFile,
		EnableWatcher: false,
	})
	if err != nil {
		t.Fatalf("Failed to create domain service: %v", err)
	}
	defer service.Close()

	// Create a new domain handler
	handler := NewDomainHandler(service)

	// Register routes
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
		resp := httptest.NewRequest("GET", "/api/v1/domains/example.com", nil)

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
		resp := httptest.NewRequest("GET", "/api/v1/domains/nonexistent.com", nil)

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
		resp := httptest.NewRequest("GET", "/api/v1/domains", nil)

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
			AlternativeNames: []string{"www.example.com", "api.example.com"},
			Enabled:          true,
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

	// Test DeleteDomain
	t.Run("DeleteDomain", func(t *testing.T) {
		resp := httptest.NewRequest("DELETE", "/api/v1/domains/example.com", nil)

		result, err := app.Test(resp)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if result.StatusCode != fiber.StatusNoContent {
			t.Errorf("Expected status %d, got %d", fiber.StatusNoContent, result.StatusCode)
		}
	})
}
