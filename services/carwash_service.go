package services

import (
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


type Slot struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Available   bool      `json:"available"`
	CurrentCars int       `json:"current_cars"`
	MaxCars     int       `json:"max_cars"`
}


type CarWashService struct {
	carwashRepository repositories.CarWashRepository
	bookingRepository repositories.BookingRepository
}

func NewCarWashService(carwashRepository repositories.CarWashRepository) *CarWashService {
	return &CarWashService{carwashRepository: carwashRepository}
}

func (cws *CarWashService) GetAvailableSlots(carwashID primitive.ObjectID, date time.Time) ([]Slot, error) {
	// objID, err := primitive.ObjectIDFromHex(carwashID)

	carwash, err := cws.carwashRepository.GetCarwashByID(carwashID)
	if err != nil {
		return nil, fmt.Errorf("carwash not found: %w", err)
	}

	bookings, err := cws.bookingRepository.GetBookingsByCarwashID(carwashID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bookings: %w", err)
	}

	const slotDuration = 30 * time.Minute
	day := date.Weekday().String()
	timeRange, exists := carwash.OpenHours[day]
	if !exists {
		return nil, errors.New("no open hours defined for the specified day")
	}

	start, err := time.Parse("15:04", timeRange.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	end, err := time.Parse("15:04", timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	start = time.Date(date.Year(), date.Month(), date.Day(), start.Hour(), start.Minute(), 0, 0, date.Location())
	end = time.Date(date.Year(), date.Month(), date.Day(), end.Hour(), end.Minute(), 0, 0, date.Location())

	var slots []Slot
	currentTime := start
	for currentTime.Before(end) {
		slotEnd := currentTime.Add(slotDuration)
		if slotEnd.After(end) {
			break
		}

		currentCars := 0
		for _, booking := range bookings {
			if booking.BookingTime.Equal(currentTime) || (booking.BookingTime.After(currentTime) && booking.BookingTime.Before(slotEnd)) {
				currentCars++
			}
		}

		slots = append(slots, Slot{
			StartTime:   currentTime,
			EndTime:     slotEnd,
			Available:   currentCars < carwash.MaxCarsPerSlot,
			CurrentCars: currentCars,
			MaxCars:     carwash.MaxCarsPerSlot,
		})

		currentTime = slotEnd
	}

	return slots, nil
}

func (cws *CarWashService) CreateCarwash(input models.Carwash) (*models.Carwash, error) {
	input.SetDefaults()

	_, err := cws.carwashRepository.CreateCarwash(input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (cws *CarWashService) GetCarwashByID(id string) (*models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid carwash ID format")
	}

	return cws.carwashRepository.GetCarwashByID(objID)
}

func (cws *CarWashService) GetAllActiveCarwashes() ([]models.Carwash, error) {
	return cws.carwashRepository.GetActiveCarwashes()
}

func (cws *CarWashService) UpdateCarwash(id string, updateData map[string]interface{}) error {
	updateData["updated_at"] = time.Now()
	return cws.carwashRepository.UpdateCarwash(id, updateData)
}

func (cws *CarWashService) SetCarwashStatus(id string, isActive bool) error {
	return cws.carwashRepository.SetCarwashStatus(id, isActive)
}

func (cws *CarWashService) CompleteOnboarding(id primitive.ObjectID) error {
	// 1. Fetch carwash details
	carwash, err := cws.carwashRepository.GetCarwashByID(id)
	if err != nil {
		return errors.New("carwash not found")
	}

	// 2. Check if already onboarded
	if carwash.HasOnboarded {
		return errors.New("carwash already completed onboarding")
	}

	// 3. Update onboarding status
	return cws.carwashRepository.Postboarding(id, true)
}



func (cws *CarWashService) GetCarwashesByOwnerID(ownerID string) ([]models.Carwash, error) {
	return cws.carwashRepository.GetCarwashesByOwnerID(ownerID)
}

func (cws *CarWashService) SearchCarwashes(filter bson.M) ([]models.Carwash, error) {
	return cws.carwashRepository.GetCarwashesByFilter(filter)
}

func (cws *CarWashService) UpdateQueueCount(carwashID string, count int) error {
	return cws.carwashRepository.UpdateQueueCount(carwashID, count)
}

func (cws *CarWashService) GetNearbyCarwashesForUser(userLat, userLng float64) (*models.CarwashSearchResult, error) {
	carwashes, searchType, err := cws.carwashRepository.FindCarwashesWithFallback(userLat, userLng)
	if err != nil {
		return nil, err
	}

	var carwashesWithDistance []models.CarwashWithDistance
	for _, carwash := range carwashes {
		distanceKm := utils.CalculateDistance(
			userLat, userLng,
			carwash.Location.Coordinates[1],
			carwash.Location.Coordinates[0],
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

	sort.Slice(carwashesWithDistance, func(i, j int) bool {
		return carwashesWithDistance[i].DistanceKm < carwashesWithDistance[j].DistanceKm
	})

	message := cws.generateSearchMessage(searchType, len(carwashesWithDistance))

	return &models.CarwashSearchResult{
		Carwashes:   carwashesWithDistance,
		SearchType:  searchType,
		UserLat:     userLat,
		UserLng:     userLng,
		ResultCount: len(carwashesWithDistance),
		Message:     message,
	}, nil
}

func (cws *CarWashService) UpdateCarwashLocation(carwashID string, locationReq models.LocationUpdateRequest) error {
	updateData := map[string]interface{}{
		"location": models.GeoLocation{
			Type:        "Point",
			Coordinates: []float64{locationReq.Longitude, locationReq.Latitude},
		},
		"service_range_minutes": locationReq.ServiceRangeMinutes,
		"has_location":          true,
		"updated_at":            time.Now(),
	}

	if locationReq.Address != "" {
		updateData["address"] = locationReq.Address
	}

	return cws.carwashRepository.UpdateCarwash(carwashID, updateData)
}

func (cws *CarWashService) generateSearchMessage(searchType string, count int) string {
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

// CreateService creates a new service for a carwash
func (cws *CarWashService) CreateService(carwashID string, service models.Service) (models.Service, error) {
	if err := service.Validate(); err != nil {
		return models.Service{}, fmt.Errorf("invalid service data: %w", err)
	}

	_, err := cws.carwashRepository.CreateService(carwashID, service)
	if err != nil {
		return models.Service{}, err
	}

	return service, nil
}

// GetServices retrieves all services for a carwash
func (cws *CarWashService) GetServices(carwashID string) ([]models.Service, error) {
	_, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID format")
	}

	return cws.carwashRepository.GetServices(carwashID)
}

// UpdateService updates an existing service for a carwash
func (cws *CarWashService) UpdateService(carwashID, serviceID string, service models.Service) error {
	if err := service.Validate(); err != nil {
		return fmt.Errorf("invalid service data: %w", err)
	}

	_, err := cws.carwashRepository.UpdateService(carwashID, serviceID, service)
	if err != nil {
		return err
	}

	return nil
}

// DeleteService deletes a service from a carwash
func (cws *CarWashService) DeleteService(carwashID, serviceID string) error {
	_, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return errors.New("invalid carwash ID format")
	}

	_, err = primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return errors.New("invalid service ID format")
	}

	_, err = cws.carwashRepository.DeleteService(carwashID, serviceID)
	if err != nil {
		return err
	}

	return nil
}

