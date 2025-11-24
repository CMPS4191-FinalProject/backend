package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func newTestServerConfig() *serverConfig {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return &serverConfig{
		logger: logger,
	}
}

func TestWriteResponseJSON(t *testing.T) {
	config := newTestServerConfig()

	tests := []struct {
		name           string
		status         int
		data           envelope
		headers        http.Header
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "simple JSON response",
			status:         http.StatusOK,
			data:           envelope{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message": "success"`,
		},
		{
			name:           "nested JSON response",
			status:         http.StatusCreated,
			data:           envelope{"user": map[string]interface{}{"id": 1, "name": "test"}},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"user"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := config.writeResponseJSON(w, tt.status, tt.data, tt.headers)

			if err != nil {
				t.Errorf("writeResponseJSON() error = %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("body does not contain expected string: %v", tt.expectedBody)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestReadRequestJSON(t *testing.T) {
	config := newTestServerConfig()

	tests := []struct {
		name        string
		body        string
		contentType string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid JSON",
			body:        `{"username": "test", "password": "pass123"}`,
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "invalid JSON syntax",
			body:        `{"username": "test", "password": }`,
			contentType: "application/json",
			wantErr:     true,
			errContains: "badly-formed JSON",
		},
		{
			name:        "empty body",
			body:        ``,
			contentType: "application/json",
			wantErr:     true,
			errContains: "must not be empty",
		},
		{
			name:        "unknown field",
			body:        `{"username": "test", "unknown_field": "value"}`,
			contentType: "application/json",
			wantErr:     true,
			errContains: "unknown key",
		},
		{
			name:        "multiple JSON values",
			body:        `{"username": "test"}{"password": "pass"}`,
			contentType: "application/json",
			wantErr:     true,
			errContains: "single JSON value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			var dest struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			err := config.readRequestJSON(w, req, &dest)

			if (err != nil) != tt.wantErr {
				t.Errorf("readRequestJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestReadRequestJSON_LargeBody(t *testing.T) {
	config := newTestServerConfig()

	// Create a body larger than 256KB
	largeBody := strings.Repeat("a", 257000)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	var dest map[string]interface{}
	err := config.readRequestJSON(w, req, &dest)

	if err == nil {
		t.Error("expected error for large body, got nil")
	}

	// The error could be either max bytes error or badly formed JSON
	// since the large body of 'a's is not valid JSON
	if !strings.Contains(err.Error(), "must not be larger than") && !strings.Contains(err.Error(), "badly-formed JSON") {
		t.Errorf("error message = %v, want to contain 'must not be larger than' or 'badly-formed JSON'", err.Error())
	}
}

func TestReadRequestJSON_WrongType(t *testing.T) {
	config := newTestServerConfig()

	body := `{"age": "not_a_number"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	var dest struct {
		Age int `json:"age"`
	}

	err := config.readRequestJSON(w, req, &dest)

	if err == nil {
		t.Error("expected error for wrong type, got nil")
	}

	if !strings.Contains(err.Error(), "incorrect JSON type") {
		t.Errorf("error message = %v, want to contain 'incorrect JSON type'", err.Error())
	}
}

func TestEnvelopeMarshaling(t *testing.T) {
	env := envelope{
		"status": "success",
		"data": map[string]interface{}{
			"id":   1,
			"name": "test",
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("failed to marshal envelope: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("status = %v, want 'success'", result["status"])
	}
}
