package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetEnvAsString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom_value",
			expected:     "custom_value",
		},
		{
			name:         "returns default when env not set",
			key:          "NONEXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvAsString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "returns env value when valid int",
			key:          "TEST_INT",
			defaultValue: 42,
			envValue:     "100",
			expected:     100,
		},
		{
			name:         "returns default when env not set",
			key:          "NONEXISTENT_INT",
			defaultValue: 42,
			envValue:     "",
			expected:     42,
		},
		{
			name:         "returns default when env value is not int",
			key:          "TEST_INVALID_INT",
			defaultValue: 42,
			envValue:     "not_an_int",
			expected:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "returns true for 'yes'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "yes",
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "returns default when env not set",
			key:          "NONEXISTENT_BOOL",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns default for invalid value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "maybe",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvAsBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "basic limit and offset",
			queryParams:    "limit=10&offset=20",
			expectedLimit:  10,
			expectedOffset: 20,
		},
		{
			name:           "limit exceeds max (capped at 100)",
			queryParams:    "limit=500&offset=0",
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "no parameters",
			queryParams:    "",
			expectedLimit:  0,
			expectedOffset: 0,
		},
		{
			name:           "page and size pagination",
			queryParams:    "page=2&size=15",
			expectedLimit:  15,
			expectedOffset: 15,
		},
		{
			name:           "page and size with size exceeding max",
			queryParams:    "page=1&size=200",
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "invalid limit value",
			queryParams:    "limit=invalid&offset=10",
			expectedLimit:  0,
			expectedOffset: 10,
		},
		{
			name:           "negative offset ignored",
			queryParams:    "limit=10&offset=-5",
			expectedLimit:  10,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.queryParams, nil)
			limit, offset := parsePaginationParams(req)

			if limit != tt.expectedLimit {
				t.Errorf("limit = %v, want %v", limit, tt.expectedLimit)
			}
			if offset != tt.expectedOffset {
				t.Errorf("offset = %v, want %v", offset, tt.expectedOffset)
			}
		})
	}
}

func TestParseSortParams(t *testing.T) {
	tests := []struct {
		name              string
		queryParams       string
		expectedSortBy    string
		expectedSortOrder string
	}{
		{
			name:              "valid sort parameters",
			queryParams:       "sort_by=username&sort_order=desc",
			expectedSortBy:    "username",
			expectedSortOrder: "desc",
		},
		{
			name:              "default sort order when invalid",
			queryParams:       "sort_by=user_id&sort_order=invalid",
			expectedSortBy:    "user_id",
			expectedSortOrder: "asc",
		},
		{
			name:              "no parameters defaults to asc",
			queryParams:       "",
			expectedSortBy:    "",
			expectedSortOrder: "asc",
		},
		{
			name:              "only sort_by provided",
			queryParams:       "sort_by=email",
			expectedSortBy:    "email",
			expectedSortOrder: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.queryParams, nil)
			sortBy, sortOrder := parseSortParams(req)

			if sortBy != tt.expectedSortBy {
				t.Errorf("sortBy = %v, want %v", sortBy, tt.expectedSortBy)
			}
			if sortOrder != tt.expectedSortOrder {
				t.Errorf("sortOrder = %v, want %v", sortOrder, tt.expectedSortOrder)
			}
		})
	}
}

func TestV(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		apiVersion  string
		expected    string
	}{
		{
			name:       "default version",
			input:      "/users",
			apiVersion: "",
			expected:   "/v1/users",
		},
		{
			name:       "custom version",
			input:      "/healthcheck",
			apiVersion: "v2",
			expected:   "/v2/healthcheck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiVersion != "" {
				os.Setenv("API_VERSION", tt.apiVersion)
				defer os.Unsetenv("API_VERSION")
			} else {
				os.Unsetenv("API_VERSION")
			}

			result := v(tt.input)
			if result != tt.expected {
				t.Errorf("v() = %v, want %v", result, tt.expected)
			}
		})
	}
}
