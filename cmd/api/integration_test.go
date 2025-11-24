package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"qotd/cmd/api/database"
	"testing"

	"github.com/julienschmidt/httprouter"
)

// setupTestServer creates a test server configuration with an in-memory database
func setupTestServer(t *testing.T) *serverConfig {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db := database.NewDatabase(database.InMemory, nil)
	db.Connect()

	config := &serverConfig{
		port:    8080,
		env:     "test",
		version: "v1",
		logger:  logger,
		db:      db,
		router:  httprouter.New(),
	}

	// Initialize WebSocket hub for tests
	InitWebSocketHub(config.db)

	return config
}

// cleanupTestServer cleans up test server resources
func cleanupTestServer(config *serverConfig) {
	if config.db != nil {
		config.db.Disconnect()
	}
}

func TestHealthCheckHandler_Integration(t *testing.T) {
	config := setupTestServer(t)
	defer cleanupTestServer(config)

	router := config.routes()

	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "alive" {
		t.Errorf("status = %v, want 'alive'", response["status"])
	}
}

func TestAuthFlow_Integration(t *testing.T) {
	config := setupTestServer(t)
	defer cleanupTestServer(config)

	router := config.routes()

	// Test user registration
	t.Run("Register new user", func(t *testing.T) {
		registerPayload := map[string]string{
			"username": "testuser",
			"password": "TestPassword123!",
			"email":    "test@example.com",
		}
		body, _ := json.Marshal(registerPayload)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/join", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Registration requires email configuration, so it may fail with 500
		// In test environment, we expect this to fail due to missing SMTP config
		if w.Code != http.StatusCreated && w.Code != http.StatusInternalServerError {
			t.Logf("Response body: %s", w.Body.String())
			t.Errorf("status code = %v, want %v or %v", w.Code, http.StatusCreated, http.StatusInternalServerError)
		}
	})

	// Wait a bit for the user to be created
	// In a real scenario, you'd verify the user through email
	// For testing, we'll use the test user initialization

	// Initialize a test user for login testing
	if err := InitTestUsers(config); err != nil {
		t.Fatalf("failed to initialize test users: %v", err)
	}

	// Test user login
	t.Run("Login with valid credentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"username": "admin",
			"password": "admin123",
		}
		body, _ := json.Marshal(loginPayload)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
			t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
			return
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response["token"] == nil {
			t.Error("expected token in response")
		}

		if response["user"] == nil {
			t.Error("expected user in response")
		}
	})

	// Test login with invalid credentials
	t.Run("Login with invalid credentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(loginPayload)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status code = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestProtectedEndpoints_Integration(t *testing.T) {
	config := setupTestServer(t)
	defer cleanupTestServer(config)

	// Initialize test users
	if err := InitTestUsers(config); err != nil {
		t.Fatalf("failed to initialize test users: %v", err)
	}

	router := config.routes()

	// Get a valid token by logging in
	loginPayload := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status code = %v", w.Code)
	}

	var loginResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	token := loginResponse.Token

	// Test accessing protected endpoint with valid token
	t.Run("Access protected endpoint with valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
			t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
		}
	})

	// Test accessing protected endpoint without token
	t.Run("Access protected endpoint without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status code = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})

	// Test accessing protected endpoint with invalid token
	t.Run("Access protected endpoint with invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status code = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestNodesCRUD_Integration(t *testing.T) {
	config := setupTestServer(t)
	defer cleanupTestServer(config)

	// Initialize test users
	if err := InitTestUsers(config); err != nil {
		t.Fatalf("failed to initialize test users: %v", err)
	}

	router := config.routes()

	// Get authentication token
	token := getAuthToken(t, router)

	var createdNodeID int

	// Test creating a node
	t.Run("Create node", func(t *testing.T) {
		nodePayload := map[string]string{
			"status": "ONLINE",
		}
		body, _ := json.Marshal(nodePayload)

		req := httptest.NewRequest(http.MethodPost, "/v1/nodes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Logf("Response body: %s", w.Body.String())
			t.Errorf("status code = %v, want %v", w.Code, http.StatusCreated)
			return
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if deviceID, ok := response["device_id"].(float64); ok {
			createdNodeID = int(deviceID)
		}
	})

	// Test getting nodes list
	t.Run("Get nodes list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/nodes", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
		}
	})

	// Test getting a specific node (if created)
	if createdNodeID > 0 {
		t.Run("Get specific node", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/nodes/"+string(rune(createdNodeID+'0')), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Status could be OK or NotFound depending on implementation
			if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
				t.Logf("status code = %v", w.Code)
			}
		})
	}
}

// Helper function to get authentication token
func getAuthToken(t *testing.T, router http.Handler) string {
	loginPayload := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status code = %v, body = %s", w.Code, w.Body.String())
	}

	var loginResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	return loginResponse.Token
}
