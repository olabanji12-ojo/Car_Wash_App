package services

import (
	// "context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateCarwash(input models.Carwash) (*models.Carwash, error) {
	
	input.SetDefaults()

	_, err := repositories.CreateCarwash(input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func GetCarwashByID(id string) (*models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid carwash ID format")
	}

	return repositories.GetCarwashByID(objID)
}

func GetAllActiveCarwashes() ([]models.Carwash, error) {
	return repositories.GetActiveCarwashes()
}

func UpdateCarwash(id string, updateData map[string]interface{}) error {
	updateData["updated_at"] = time.Now()
	return repositories.UpdateCarwash(id, updateData)
}

func SetCarwashStatus(id string, isActive bool) error {

	return repositories.SetCarwashStatus(id, isActive)

}

func GetCarwashesByOwnerID(ownerID string) ([]models.Carwash, error) {

	return repositories.GetCarwashesByOwnerID(ownerID)

}

func SearchCarwashes(filter bson.M) ([]models.Carwash, error) {

	return repositories.GetCarwashesByFilter(filter)

}

func UpdateQueueCount(carwashID string, count int) error {
	return repositories.UpdateQueueCount(carwashID, count)
}

// GetNearbyCarwashesForUser finds carwashes near user location with fallback strategy
func GetNearbyCarwashesForUser(userLat, userLng float64) (*models.CarwashSearchResult, error) {
	carwashes, searchType, err := repositories.FindCarwashesWithFallback(userLat, userLng)
	if err != nil {
		return nil, err
	}

	// Convert to CarwashWithDistance and calculate distances
	var carwashesWithDistance []models.CarwashWithDistance
	for _, carwash := range carwashes {
		distanceKm := utils.CalculateDistance(
			userLat, userLng,
			carwash.Location.Coordinates[1], // latitude
			carwash.Location.Coordinates[0], // longitude
		)
		estimatedTravelTime := utils.EstimateTravelTimeMinutes(distanceKm)
		isWithinRange := utils.IsWithinServiceRange(
			userLat, userLng,
			carwash.Location.Coordinates[1],
			carwash.Location.Coordinates[0],
			carwash.ServiceRangeMinutes,
			
		)

		carwashesWithDistance = append(carwashesWithDistance, models.CarwashWithDistance{
			Carwash:              *carwash,
			DistanceKm:           distanceKm,
			EstimatedTravelTime:  estimatedTravelTime,
			IsWithinServiceRange: isWithinRange,
		})
	}

	// Sort by distance (closest first)
	sort.Slice(carwashesWithDistance, func(i, j int) bool {
		return carwashesWithDistance[i].DistanceKm < carwashesWithDistance[j].DistanceKm
	})

	// Generate appropriate message based on search type
	message := generateSearchMessage(searchType, len(carwashesWithDistance))

	return &models.CarwashSearchResult{
		Carwashes:   carwashesWithDistance,
		SearchType:  searchType,
		UserLat:     userLat,
		UserLng:     userLng,
		ResultCount: len(carwashesWithDistance),
		Message:     message,
	}, nil
}

// UpdateCarwashLocation updates carwash location and service range
func UpdateCarwashLocation(carwashID string, locationReq models.LocationUpdateRequest) error {
	updateData := map[string]interface{}{
		"location": models.GeoLocation{
			Type:        "Point",
			Coordinates: []float64{locationReq.Longitude, locationReq.Latitude},
		},
		"service_range_minutes": locationReq.ServiceRangeMinutes,
		"has_location":          true,
		"updated_at":           time.Now(),
	}

	if locationReq.Address != "" {
		updateData["address"] = locationReq.Address
	}

	return UpdateCarwash(carwashID, updateData)
}

// Helper function to generate search result messages
func generateSearchMessage(searchType string, count int) string {
	switch searchType {
	case "nearby":
		return fmt.Sprintf("Found %d carwashes within 10km of your location", count)
	case "extended":
		return fmt.Sprintf("Found %d carwashes within 100km of your location", count)
	case "all":
		return fmt.Sprintf("Showing all %d available carwashes with location data", count)
	default:
		return fmt.Sprintf("Found %d carwashes", count)
	}
}