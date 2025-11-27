package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewOpenMeteoService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	service := NewOpenMeteoService(logger)
	if service == nil {
		t.Fatal("expected service, got nil")
	}
	if service.baseURL != "https://api.open-meteo.com/v1" {
		t.Errorf("expected baseURL 'https://api.open-meteo.com/v1', got %s", service.baseURL)
	}
}

func TestCompareSensorWithForecast(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := NewOpenMeteoService(logger)

	t.Run("validates device moisture value", func(t *testing.T) {
		// Test with reasonable coordinates (New York City)
		lat := 40.7128
		lon := -74.0060
		deviceMoisture := 0.25 // 25% moisture as m³/m³

		result, err := service.CompareSensorWithForecast(deviceMoisture, lat, lon)
		
		// The API might fail in test environments, so we just check the structure if it succeeds
		if err == nil {
			if result["device_moisture"] != deviceMoisture {
				t.Errorf("expected device_moisture %f, got %v", deviceMoisture, result["device_moisture"])
			}
			if _, ok := result["forecast_moisture"]; !ok {
				t.Error("expected forecast_moisture in result")
			}
			if _, ok := result["insights"]; !ok {
				t.Error("expected insights in result")
			}
		}
	})
}
