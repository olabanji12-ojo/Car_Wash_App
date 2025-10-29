// services/geocoding/geocoder.go
package geocoding

import "context"

// Location represents a geographic point with latitude and longitude
type Location struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

// Address represents a structured address
type Address struct {
    FormattedAddress string            `json:"formatted_address"`
    Components       map[string]string `json:"components"`
}

// Geocoder defines the interface for geocoding operations
type Geocoder interface {
    // Geocode converts an address to geographic coordinates
    Geocode(ctx context.Context, address string) (*Location, error)
    // ReverseGeocode converts coordinates to an address
    ReverseGeocode(ctx context.Context, lat, lng float64) (*Address, error)
}