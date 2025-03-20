package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

func TestDomainHandler(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Create a new domain service
	ds, err := service.NewDomainService(domainsFile)
	if err != nil {
		t.Fatalf("Failed to create domain service: %v", err)
	}

	// Create a new domain handler
	dh := NewDomainHandler(ds)

	// Create a new Fiber app
	app := fiber.New()
	dh.RegisterRoutes(app)

	// Test data
	testDomain := model.CreateDomainRequest{
		Domain:           "example.com",
		AlternativeNames: []string{"www.example.com"},
		Enabled:          true,
		Comment:          "Test comment",
	}

	// Test creating a domain
	t.Run("CreateDomain", func(t *testing.T) {
		body, _ := json.Marshal(testDomain)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		var response model.DomainResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data.Domain != testDomain.Domain {
			t.Errorf("Expected domain %s, got %s", testDomain.Domain, response.Data.Domain)
		}
	})

	// Test creating an invalid domain
	t.Run("CreateInvalidDomain", func(t *testing.T) {
		invalidDomain := model.CreateDomainRequest{
			Domain:           "invalid@domain.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}
		body, _ := json.Marshal(invalidDomain)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	// Test getting a domain
	t.Run("GetDomain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/example.com", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response model.DomainResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data.Domain != testDomain.Domain {
			t.Errorf("Expected domain %s, got %s", testDomain.Domain, response.Data.Domain)
		}
	})

	// Test getting a non-existent domain
	t.Run("GetNonExistentDomain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/nonexistent.com", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	// Test listing domains
	t.Run("ListDomains", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response model.DomainsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if !response.Success {
			t.Error("Expected success to be true")
		}
		if len(response.Data) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(response.Data))
		}
	})

	// Test updating a domain
	t.Run("UpdateDomain", func(t *testing.T) {
		updateReq := model.UpdateDomainRequest{
			AlternativeNames: []string{"www.example.com", "mail.example.com"},
			Enabled:          true,
		}
		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/domains/example.com", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response model.DomainResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if !response.Success {
			t.Error("Expected success to be true")
		}
		if len(response.Data.AlternativeNames) != 2 {
			t.Errorf("Expected 2 alternative names, got %d", len(response.Data.AlternativeNames))
		}
	})

	// Test deleting a domain
	t.Run("DeleteDomain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/example.com", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
		}

		// Verify the domain was deleted
		req = httptest.NewRequest(http.MethodGet, "/api/v1/domains/example.com", nil)
		resp, err = app.Test(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}
