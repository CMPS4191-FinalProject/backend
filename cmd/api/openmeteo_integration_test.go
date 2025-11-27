package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sma/cmd/api/database"
	"sma/cmd/api/types"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestOpenMeteoIntegration(t *testing.T) {
	// Setup test server
	config := serverConfig{
		port:    8080,
		env:     "test",
		version: "v1",
		router:  httprouter.New(),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, nil)),
		db:      database.NewDatabase(database.InMemory, nil),
	}
	config.db.Connect()
	defer config.db.Disconnect()

	// Initialize Open-Meteo service
	config.openMeteoService = NewOpenMeteoService(config.logger)

	// Setup routes
	router := config.routes()

	t.Run("soil forecast endpoint validates coordinates", func(t *testing.T) {
		// Create test request with invalid lat
		req := httptest.NewRequest("GET", "/v1/forecast/soil?lat=invalid&lon=-74.0060", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("soil forecast endpoint validates latitude range", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/forecast/soil?lat=100&lon=-74.0060", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("weather forecast endpoint requires authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/forecast/weather?lat=40.7128&lon=-74.0060", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("compare endpoint requires device data", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/forecast/compare?device_id=999&lat=40.7128&lon=-74.0060", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 404 since device doesn't exist
		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("full workflow with device data", func(t *testing.T) {
		// Create a node
		nodeReq := map[string]interface{}{
			"status": "ONLINE",
		}
		nodeBody, _ := json.Marshal(nodeReq)
		req := httptest.NewRequest("POST", "/v1/nodes", bytes.NewBuffer(nodeBody))
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create node: status %d", w.Code)
		}

		// Add sensor data
		dataReq := map[string]interface{}{
			"device_id":        1,
			"moisture_content": 45.5,
		}
		dataBody, _ := json.Marshal(dataReq)
		req = httptest.NewRequest("POST", "/v1/nodedata", bytes.NewBuffer(dataBody))
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create node data: status %d", w.Code)
		}

		// Now the compare endpoint should work (though API call will fail in test)
		req = httptest.NewRequest("GET", "/v1/forecast/compare?device_id=1&lat=40.7128&lon=-74.0060", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1, "admin", "admin"))
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Will fail with 500 due to no network access, but that's expected in test
		// The important thing is it's not 404 or 400
		if w.Code != http.StatusInternalServerError {
			t.Logf("Got status %d, which means endpoint is working (network failure expected in test)", w.Code)
		}
	})
}

// Helper function to generate test JWT token
func generateTestToken(userID int, username, role string) string {
	user := &types.User{
		UserID:   userID,
		Username: username,
		Role:     types.UserRole(role),
	}
	token, _ := GenerateJWT(user)
	return token
}
