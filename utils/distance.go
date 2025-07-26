package utils

import (

	"math"

)

// CalculateDistance returns distance in kilometers between two geo coordinates
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return distance
}


// EstimateTravelTimeMinutes estimates travel time in minutes based on distance
// Uses average city driving speed of 30 km/h
func EstimateTravelTimeMinutes(distanceKm float64) int {
	const avgSpeedKmh = 30.0 // Average city driving speed
	travelTimeHours := distanceKm / avgSpeedKmh
	travelTimeMinutes := int(math.Ceil(travelTimeHours * 60))
	return travelTimeMinutes
}

// IsWithinServiceRange checks if user location is within carwash service range
func IsWithinServiceRange(userLat, userLng, carwashLat, carwashLng float64, serviceRangeMinutes int) bool {
	distanceKm := CalculateDistance(userLat, userLng, carwashLat, carwashLng)
	estimatedTravelTime := EstimateTravelTimeMinutes(distanceKm)
	return estimatedTravelTime <= serviceRangeMinutes
}


