package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

func v(str string) string {
	return "/" + getEnvAsString("API_VERSION", "v1") + str
}

func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		// Convert to lowercase for case-insensitive comparison
		lowerValue := strings.ToLower(value)
		if lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" {
			return true
		} else if lowerValue == "false" || lowerValue == "0" || lowerValue == "no" {
			return false
		}
	}
	return defaultValue
}

// parsePaginationParams extracts pagination parameters from query string
func parsePaginationParams(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 { // Cap maximum limit
				limit = 100
			}
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Alternative pagination using page and size
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")
	if pageStr != "" && sizeStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
				if size > 100 { // Cap maximum size
					size = 100
				}
				limit = size
				offset = (page - 1) * size
			}
		}
	}

	return limit, offset
}

// parseSortParams extracts sorting parameters from query string
func parseSortParams(r *http.Request) (sortBy, sortOrder string) {
	sortBy = r.URL.Query().Get("sort_by")
	sortOrder = r.URL.Query().Get("sort_order")

	// Normalize sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // default
	}

	return sortBy, sortOrder
}
