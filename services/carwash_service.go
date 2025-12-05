package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/services/geocoding"
	"github.com/sirupsen/logrus"
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
	geocoder          geocoding.Geocoder
}

func NewCarWashService(
	carwashRepository repositories.CarWashRepository,
	bookingRepository repositories.BookingRepository,
	geocoder geocoding.Geocoder,
) *CarWashService {
	return &CarWashService{
		carwashRepository: carwashRepository,
		bookingRepository: bookingRepository,
		geocoder:          geocoder,
	}
}

func (cws *CarWashService) GetAvailableSlots(carwashID primitive.ObjectID, date time.Time) ([]Slot, error) {
	carwash, err := cws.carwashRepository.GetCarwashByID(carwashID)
	if err != nil {
		return nil, fmt.Errorf("carwash not found: %w", err)
	}

	bookings, err := cws.bookingRepository.GetBookingsByCarwashID(carwashID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bookings: %w", err)
	}

	const slotDuration = 30 * time.Minute
	day := strings.ToLower(date.Weekday().String())
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

	// If we have an address but no coordinates, geocode it
	if input.Address != "" && (input.Location.Coordinates == nil || len(input.Location.Coordinates) == 0) {
		logrus.Infof("Geocoding address: %s", input.Address)
		location, err := cws.geocoder.Geocode(context.Background(), input.Address)
		if err != nil {
			logrus.Warnf("Failed to geocode address '%s': %v", input.Address, err)
			// Continue without failing, as coordinates might be optional
		} else {
			input.Location = models.GeoLocation{
				Type:        "Point",
				Coordinates: []float64{location.Lng, location.Lat}, // GeoJSON uses [lng, lat] order
			}
			input.HasLocation = true
			logrus.Infof("Successfully geocoded address to coordinates: [%f, %f]", location.Lng, location.Lat)
		}
	}

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
	// If address is being updated, geocode the new address
	if address, ok := updateData["address"].(string); ok && address != "" {
		logrus.Infof("Geocoding updated address: %s", address)
		location, err := cws.geocoder.Geocode(context.Background(), address)
		if err != nil {
			logrus.Warnf("Failed to geocode updated address '%s': %v", address, err)
		} else {
			// Update the location in the update data
			updateData["location"] = models.GeoLocation{
				Type:        "Point",
				Coordinates: []float64{location.Lng, location.Lat}, // GeoJSON uses [lng, lat] order
			}
			updateData["has_location"] = true
			logrus.Infof("Successfully geocoded updated address to coordinates: [%f, %f]", location.Lng, location.Lat)
		}
	}

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

// GetNearbyCarwashesForUser finds car washes near the user's location
// Returns a map with carwashes including distance information
func (cws *CarWashService) GetNearbyCarwashesForUser(userLat, userLng float64) (map[string]interface{}, error) {
	// Get carwashes with distance information from the repository
	carwashes, searchType, err := cws.carwashRepository.FindCarwashesWithFallback(userLat, userLng)
	if err != nil {
		return nil, fmt.Errorf("error finding nearby carwashes: %w", err)
	}

	// Generate a user-friendly message about the search results
	message := cws.generateSearchMessage(searchType, len(carwashes))

	// Prepare the response
	response := map[string]interface{}{
		"carwashes":   carwashes,
		"search_type": searchType,
		"user_lat":    userLat,
		"user_lng":    userLng,
		"count":       len(carwashes),
		"message":     message,
	}

	return response, nil
}

// generateSearchMessage creates a user-friendly message about the search results
func (cws *CarWashService) generateSearchMessage(searchType string, count int) string {
	switch searchType {
	case "nearby":
		if count == 0 {
			return "No car washes found nearby. Expanding search radius..."
		}
		return fmt.Sprintf("Found %d car washes within 10km", count)
	case "extended":
		if count == 0 {
			return "No car washes found in the extended area. Showing all available locations..."
		}
		return fmt.Sprintf("Found %d car washes within 100km", count)
	case "all":
		if count == 0 {
			return "No car washes available at this time."
		}
		return fmt.Sprintf("Showing all %d available car washes", count)
	default:
		return fmt.Sprintf("Found %d car washes", count)
	}
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

// CreateService creates a new service for a carwash
func (cws *CarWashService) CreateService(carwashID string, service models.Service) (models.Service, error) {
	// Validate service input
	if err := service.Validate(); err != nil {
		logrus.Errorf("Invalid service data: %v", err)
		return models.Service{}, fmt.Errorf("invalid service data: %w", err)
	}

	// Convert carwashID to ObjectID
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Errorf("Invalid carwash ID: %v", err)
		return models.Service{}, errors.New("invalid carwash ID format")
	}

	// Fetch the carwash to check its existence and services field
	carwash, err := cws.carwashRepository.GetCarwashByID(objID)
	if err != nil {
		logrus.Errorf("Carwash not found: %v", err)
		return models.Service{}, fmt.Errorf("carwash not found: %w", err)
	}

	// Ensure services field is initialized
	if carwash.Services == nil {
		logrus.Warnf("Services field is null for carwash %s, initializing to empty array", carwashID)
		carwash := models.Carwash{
			Services: []models.Service{},
		}
		fmt.Print(carwash)
	}

	// Assign a new ObjectID to the service
	service.ID = primitive.NewObjectID()

	// Add the service to the carwash
	result, err := cws.carwashRepository.CreateService(carwashID, service)
	if err != nil {
		logrus.Errorf("Failed to create service for carwash %s: %v", carwashID, err)
		return models.Service{}, fmt.Errorf("failed to create service: %w", err)
	}

	if result.ModifiedCount == 0 {
		logrus.Errorf("No carwash found or service not added for carwash %s", carwashID)
		return models.Service{}, errors.New("no carwash found or service not added")
	}

	logrus.Infof("Successfully created service %s for carwash %s", service.Name, carwashID)
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

func (cws *CarWashService) GetServiceByID(carwashID, serviceID string) (*models.Service, error) {
	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Errorf("Invalid carwash ID: %v", err)
		return nil, errors.New("invalid carwash ID format")
	}

	serviceObjID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		logrus.Errorf("Invalid service ID: %v", err)
		return nil, errors.New("invalid service ID format")
	}

	carwash, err := cws.carwashRepository.GetCarwashByID(carwashObjID)
	if err != nil {
		logrus.Errorf("Carwash not found: %v", err)
		return nil, fmt.Errorf("carwash not found: %w", err)
	}

	for _, service := range carwash.Services {
		if service.ID == serviceObjID {
			return &service, nil
		}
	}

	logrus.Errorf("Service %s not found in carwash %s", serviceID, carwashID)
	return nil, errors.New("service not found")
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

// UploadCarwashPhoto uploads a photo for a carwash and adds it to the gallery
func (cws *CarWashService) UploadCarwashPhoto(carwashID string, photoFile *ProfilePhotoFile) (string, error) {
	// 1. Validate ID
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return "", errors.New("invalid carwash ID format")
	}

	// 2. Validate file
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if photoFile.Size > maxSize {
		return "", errors.New("file size too large (max 5MB)")
	}

	ext := strings.ToLower(filepath.Ext(photoFile.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	if !allowedExts[ext] {
		return "", errors.New("invalid file type (allowed: jpg, jpeg, png, gif, webp)")
	}

	// 3. Upload to Cloudinary
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("carwash_%s_%d%s", carwashID, timestamp, ext)

	uploadResult, err := UploadImage(photoFile.File, filename, "carwash_photos")
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %v", err)
	}

	// 4. Update DB
	err = cws.carwashRepository.AddPhotoToGallery(objID, uploadResult.SecureURL)
	if err != nil {
		// Cleanup
		DeleteImage(uploadResult.PublicID)
		return "", err
	}

	return uploadResult.SecureURL, nil
}
