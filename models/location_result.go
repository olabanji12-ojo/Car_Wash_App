package models

import (
	"errors"
)

// CarwashSearchResult represents the result of a location-based carwash search
type CarwashSearchResult struct {
	Carwashes   []CarwashWithDistance `json:"carwashes"`
	SearchType  string                `json:"search_type"`  // "nearby", "extended", "all"
	UserLat     float64               `json:"user_lat"`
	UserLng     float64               `json:"user_lng"`
	ResultCount int                   `json:"result_count"`
	Message     string                `json:"message"`
}

// CarwashWithDistance extends Carwash with distance and travel time info
type CarwashWithDistance struct {
	Carwash              `json:"carwash"`
	DistanceKm           float64 `json:"distance_km"`
	EstimatedTravelTime  int     `json:"estimated_travel_time_minutes"`
	IsWithinServiceRange bool    `json:"is_within_service_range"`
}


// LocationUpdateRequest for updating carwash location
type LocationUpdateRequest struct {
	Latitude            float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude           float64 `json:"longitude" validate:"required,min=-180,max=180"`
	ServiceRangeMinutes int     `json:"service_range_minutes" validate:"required,min=1,max=180"`
	Address             string  `json:"address,omitempty"`
}

// Validate validates the LocationUpdateRequest
func (req *LocationUpdateRequest) Validate() error {
	if req.Latitude < -90 || req.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if req.ServiceRangeMinutes < 1 || req.ServiceRangeMinutes > 180 {
		return errors.New("service range must be between 1 and 180 minutes")
	}
	return nil
}

