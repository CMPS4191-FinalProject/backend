package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OpenMeteoService handles fetching weather and soil data from Open-Meteo API
// Open-Meteo is a free, open-source weather API that requires no API key
type OpenMeteoService struct {
	baseURL string
	logger  *slog.Logger
}

// SoilMoistureData represents soil moisture forecast data from Open-Meteo
type SoilMoistureData struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Time      []string  `json:"time"`
	Moisture0to1cm   []float64 `json:"soil_moisture_0_to_1cm"`
	Moisture1to3cm   []float64 `json:"soil_moisture_1_to_3cm"`
	Moisture3to9cm   []float64 `json:"soil_moisture_3_to_9cm"`
	Moisture9to27cm  []float64 `json:"soil_moisture_9_to_27cm"`
	Moisture27to81cm []float64 `json:"soil_moisture_27_to_81cm"`
}

// WeatherForecast represents weather forecast data from Open-Meteo
type WeatherForecast struct {
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timezone    string    `json:"timezone"`
	Time        []string  `json:"time"`
	Temperature []float64 `json:"temperature_2m"`
	Humidity    []float64 `json:"relative_humidity_2m"`
	Precipitation []float64 `json:"precipitation"`
	SoilTemperature0cm []float64 `json:"soil_temperature_0cm"`
	SoilTemperature6cm []float64 `json:"soil_temperature_6cm"`
}

// OpenMeteoResponse represents the raw API response
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Hourly    struct {
		Time                  []string  `json:"time"`
		Temperature2m         []float64 `json:"temperature_2m"`
		RelativeHumidity2m    []float64 `json:"relative_humidity_2m"`
		Precipitation         []float64 `json:"precipitation"`
		SoilTemperature0cm    []float64 `json:"soil_temperature_0cm"`
		SoilTemperature6cm    []float64 `json:"soil_temperature_6cm"`
		SoilMoisture0to1cm    []float64 `json:"soil_moisture_0_to_1cm"`
		SoilMoisture1to3cm    []float64 `json:"soil_moisture_1_to_3cm"`
		SoilMoisture3to9cm    []float64 `json:"soil_moisture_3_to_9cm"`
		SoilMoisture9to27cm   []float64 `json:"soil_moisture_9_to_27cm"`
		SoilMoisture27to81cm  []float64 `json:"soil_moisture_27_to_81cm"`
	} `json:"hourly"`
}

// NewOpenMeteoService creates a new Open-Meteo service instance
func NewOpenMeteoService(logger *slog.Logger) *OpenMeteoService {
	return &OpenMeteoService{
		baseURL: "https://api.open-meteo.com/v1",
		logger:  logger,
	}
}

// GetSoilMoistureForecast fetches soil moisture forecast for given coordinates
func (oms *OpenMeteoService) GetSoilMoistureForecast(lat, lon float64, days int) (*SoilMoistureData, error) {
	if days < 1 || days > 16 {
		days = 7 // Default to 7 days
	}

	url := fmt.Sprintf("%s/forecast?latitude=%.6f&longitude=%.6f&hourly=soil_moisture_0_to_1cm,soil_moisture_1_to_3cm,soil_moisture_3_to_9cm,soil_moisture_9_to_27cm,soil_moisture_27_to_81cm&forecast_days=%d",
		oms.baseURL, lat, lon, days)

	resp, err := http.Get(url)
	if err != nil {
		oms.logger.Error("failed to fetch soil moisture data", "error", err)
		return nil, fmt.Errorf("failed to fetch soil moisture data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		oms.logger.Error("open-meteo API error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse OpenMeteoResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		oms.logger.Error("failed to parse response", "error", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	soilData := &SoilMoistureData{
		Latitude:         apiResponse.Latitude,
		Longitude:        apiResponse.Longitude,
		Timezone:         apiResponse.Timezone,
		Time:             apiResponse.Hourly.Time,
		Moisture0to1cm:   apiResponse.Hourly.SoilMoisture0to1cm,
		Moisture1to3cm:   apiResponse.Hourly.SoilMoisture1to3cm,
		Moisture3to9cm:   apiResponse.Hourly.SoilMoisture3to9cm,
		Moisture9to27cm:  apiResponse.Hourly.SoilMoisture9to27cm,
		Moisture27to81cm: apiResponse.Hourly.SoilMoisture27to81cm,
	}

	oms.logger.Info("soil moisture forecast fetched", "lat", lat, "lon", lon, "days", days)
	return soilData, nil
}

// GetWeatherForecast fetches weather and soil temperature forecast
func (oms *OpenMeteoService) GetWeatherForecast(lat, lon float64, days int) (*WeatherForecast, error) {
	if days < 1 || days > 16 {
		days = 7
	}

	url := fmt.Sprintf("%s/forecast?latitude=%.6f&longitude=%.6f&hourly=temperature_2m,relative_humidity_2m,precipitation,soil_temperature_0cm,soil_temperature_6cm&forecast_days=%d",
		oms.baseURL, lat, lon, days)

	resp, err := http.Get(url)
	if err != nil {
		oms.logger.Error("failed to fetch weather data", "error", err)
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		oms.logger.Error("open-meteo API error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse OpenMeteoResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		oms.logger.Error("failed to parse response", "error", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	forecast := &WeatherForecast{
		Latitude:        apiResponse.Latitude,
		Longitude:       apiResponse.Longitude,
		Timezone:        apiResponse.Timezone,
		Time:            apiResponse.Hourly.Time,
		Temperature:     apiResponse.Hourly.Temperature2m,
		Humidity:        apiResponse.Hourly.RelativeHumidity2m,
		Precipitation:   apiResponse.Hourly.Precipitation,
		SoilTemperature0cm: apiResponse.Hourly.SoilTemperature0cm,
		SoilTemperature6cm: apiResponse.Hourly.SoilTemperature6cm,
	}

	oms.logger.Info("weather forecast fetched", "lat", lat, "lon", lon, "days", days)
	return forecast, nil
}

// CompareSensorWithForecast compares device sensor data with Open-Meteo forecast
func (oms *OpenMeteoService) CompareSensorWithForecast(deviceMoisture float64, lat, lon float64) (map[string]interface{}, error) {
	forecast, err := oms.GetSoilMoistureForecast(lat, lon, 1)
	if err != nil {
		return nil, err
	}

	if len(forecast.Time) == 0 || len(forecast.Moisture3to9cm) == 0 {
		return nil, fmt.Errorf("no forecast data available")
	}

	// Get current forecast values (first entry)
	currentTime, _ := time.Parse(time.RFC3339, forecast.Time[0])
	forecastMoisture := forecast.Moisture3to9cm[0]

	// Calculate difference
	difference := deviceMoisture - forecastMoisture
	var percentDiff float64
	if forecastMoisture > 0.001 { // Avoid division by zero
		percentDiff = (difference / forecastMoisture) * 100
	} else {
		percentDiff = 0 // If forecast is near zero, can't calculate percentage
	}

	insights := []string{}
	
	if percentDiff > 20 {
		insights = append(insights, fmt.Sprintf("Your sensor reads %.1f%% higher moisture than forecast. Soil may be over-irrigated.", percentDiff))
	} else if percentDiff < -20 {
		insights = append(insights, fmt.Sprintf("Your sensor reads %.1f%% lower moisture than forecast. Consider irrigation.", -percentDiff))
	} else {
		insights = append(insights, "Sensor readings align well with forecast data.")
	}

	// Add forecast trend
	if len(forecast.Moisture3to9cm) > 24 {
		avgNext24h := 0.0
		for i := 0; i < 24 && i < len(forecast.Moisture3to9cm); i++ {
			avgNext24h += forecast.Moisture3to9cm[i]
		}
		avgNext24h /= float64(min(24, len(forecast.Moisture3to9cm)))
		
		trend := avgNext24h - forecastMoisture
		if trend > 0.02 {
			insights = append(insights, fmt.Sprintf("Forecast shows increasing soil moisture (avg +%.2f m³/m³ next 24h).", trend))
		} else if trend < -0.02 {
			insights = append(insights, fmt.Sprintf("Forecast shows decreasing soil moisture (avg %.2f m³/m³ next 24h).", trend))
		}
	}

	result := map[string]interface{}{
		"device_moisture":    deviceMoisture,
		"forecast_moisture":  forecastMoisture,
		"difference":         difference,
		"percent_difference": percentDiff,
		"forecast_time":      currentTime,
		"insights":           insights,
		"location": map[string]float64{
			"latitude":  forecast.Latitude,
			"longitude": forecast.Longitude,
		},
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
