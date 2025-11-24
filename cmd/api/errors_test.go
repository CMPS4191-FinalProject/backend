package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestErrorResponseFormat verifies that all error responses return consistent JSON format
func TestErrorResponseFormat(t *testing.T) {
	// Create a test server config with a real logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &serverConfig{
		logger: logger,
	}

	tests := []struct {
		name           string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Bad Request Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.badRequestResponse(w, r, "Invalid input")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid input",
		},
		{
			name: "Unauthorized Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.unauthorizedResponse(w, r, "Not authenticated")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Not authenticated",
		},
		{
			name: "Forbidden Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.forbiddenResponse(w, r, "Access denied")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
		{
			name: "Not Found Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.notFoundResponse(w, r)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  ERROR_NOTFOUND,
		},
		{
			name: "Internal Server Error Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.internalServerErrorResponse(w, r, "Database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Database error",
		},
		{
			name: "Too Many Requests Response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				config.tooManyRequestsResponse(w, r, "Rate limit exceeded")
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			tt.handlerFunc(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check content type
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want application/json",
					contentType)
			}

			// Parse the response body
			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}

			// Check that the response has an "error" field
			errorMsg, ok := response["error"]
			if !ok {
				t.Errorf("Response missing 'error' field. Got: %v", response)
			}

			// Check the error message
			if errorMsg != tt.expectedError {
				t.Errorf("Wrong error message: got %v want %v", errorMsg, tt.expectedError)
			}
		})
	}
}
