package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Handles reverse geocoding operations
type GeocodingService struct {
	client *http.Client
	config *Config
}

// Creates a new geocoding service
func NewGeocodingService(config *Config) *GeocodingService {
	return &GeocodingService{
		client: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		config: config,
	}
}

// Performs reverse geocoding using Nominatim API
func (g *GeocodingService) ReverseGeocode(lat, lon float64) (*NominatimResponse, error) {
	url := fmt.Sprintf(
		"%s?format=json&lat=%f&lon=%f&zoom=18&addressdetails=1",
		g.config.NominatimURL, lat, lon,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result NominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// Performs reverse geocoding with rate limiting
func (g *GeocodingService) ReverseGeocodeWithRateLimit(lat, lon float64) (*NominatimResponse, error) {
	location, err := g.ReverseGeocode(lat, lon)
	if err != nil {
		return nil, err
	}

	// Add delay to respect Nominatim rate limits
	time.Sleep(1 * time.Second)

	return location, nil
}
