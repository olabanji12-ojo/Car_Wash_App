// services/geocoding/google/geocoder.go
package google

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"

    "github.com/olabanji12-ojo/CarWashApp/services/geocoding"
)

// GoogleMapsGeocoder implements the Geocoder interface using Google Maps API
type GoogleMapsGeocoder struct {
    apiKey string
    client *http.Client
    baseURL string
}

// NewGoogleMapsGeocoder creates a new Google Maps geocoder instance
func NewGoogleMapsGeocoder(apiKey string) *GoogleMapsGeocoder {
    return &GoogleMapsGeocoder{
        apiKey:  apiKey,
        client:  &http.Client{Timeout: 10 * time.Second},
        baseURL: "https://maps.googleapis.com/maps/api/geocode/json",
    }
}

// GeocodeResponse represents the Google Maps Geocoding API response
type geocodeResponse struct {
    Results []struct {
        Geometry struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`
            } `json:"location"`
        } `json:"geometry"`
        FormattedAddress string `json:"formatted_address"`
        AddressComponents []struct {
            Types   []string `json:"types"`
            LongName string  `json:"long_name"`
        } `json:"address_components"`
    } `json:"results"`
    Status string `json:"status"`
}

// Geocode converts an address to geographic coordinates
func (g *GoogleMapsGeocoder) Geocode(ctx context.Context, address string) (*geocoding.Location, error) {
    // Build the request URL
    params := url.Values{}
    params.Add("address", address)
    params.Add("key", g.apiKey)
    
    reqURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Send request
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("geocoding request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Parse response
    var result geocodeResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Check API status
    if result.Status != "OK" || len(result.Results) == 0 {
        return nil, fmt.Errorf("geocoding failed with status: %s", result.Status)
    }
    
    // Return the first result
    loc := result.Results[0].Geometry.Location
    return &geocoding.Location{
        Lat: loc.Lat,
        Lng: loc.Lng,
    }, nil
}

// ReverseGeocode converts coordinates to an address
func (g *GoogleMapsGeocoder) ReverseGeocode(ctx context.Context, lat, lng float64) (*geocoding.Address, error) {
    // Build the request URL
    params := url.Values{}
    params.Add("latlng", fmt.Sprintf("%f,%f", lat, lng))
    params.Add("key", g.apiKey)
    
    reqURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Send request
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("reverse geocoding request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Parse response
    var result struct {
        Results []struct {
            FormattedAddress string `json:"formatted_address"`
            AddressComponents []struct {
                Types   []string `json:"types"`
                LongName string  `json:"long_name"`
            } `json:"address_components"`
        } `json:"results"`
        Status string `json:"status"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Check API status
    if result.Status != "OK" || len(result.Results) == 0 {
        return nil, fmt.Errorf("reverse geocoding failed with status: %s", result.Status)
    }
    
    // Extract address components
    address := &geocoding.Address{
        FormattedAddress: result.Results[0].FormattedAddress,
        Components:       make(map[string]string),
    }
    
    // Map common address components
    for _, comp := range result.Results[0].AddressComponents {
        for _, t := range comp.Types {
            switch t {
            case "street_number":
                address.Components["street_number"] = comp.LongName
            case "route":
                address.Components["route"] = comp.LongName
            case "locality":
                address.Components["city"] = comp.LongName
            case "administrative_area_level_1":
                address.Components["state"] = comp.LongName
            case "country":
                address.Components["country"] = comp.LongName
            case "postal_code":
                address.Components["postal_code"] = comp.LongName
            }
        }
    }
    
    return address, nil
}