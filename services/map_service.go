package services

import (
	"github.com/olabanji12-ojo/CarWashApp/utils"
)

// MapService defines the interface for all distance and travel time calculations.
// This allows us to switch from the current manual calculation to Google Maps or Mapbox later.
type MapService interface {
	GetDistanceKm(lat1, lon1, lat2, lon2 float64) float64
	EstimateTravelTime(distanceKm float64) int
	IsWithinRange(lat1, lon1, lat2, lon2 float64, maxKm float64) bool
}

// DefaultMapService implements MapService using the manual Haversine formula.
type DefaultMapService struct{}

func NewDefaultMapService() MapService {
	return &DefaultMapService{}
}

func (s *DefaultMapService) GetDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	return utils.CalculateDistance(lat1, lon1, lat2, lon2)
}

func (s *DefaultMapService) EstimateTravelTime(distanceKm float64) int {
	return utils.EstimateTravelTimeMinutes(distanceKm)
}

func (s *DefaultMapService) IsWithinRange(lat1, lon1, lat2, lon2 float64, maxKm float64) bool {
	distance := s.GetDistanceKm(lat1, lon1, lat2, lon2)
	return distance <= maxKm
}
