package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"strings"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingService struct {
	bookingRepository   repositories.BookingRepository
	carWashRepository   repositories.CarWashRepository
	userRepository      repositories.UserRepository
	notificationService *NotificationService
}

func NewBookingService(bookingRepository repositories.BookingRepository, carWashRepository repositories.CarWashRepository, userRepository repositories.UserRepository, notificationService *NotificationService) *BookingService {
	return &BookingService{
		bookingRepository:   bookingRepository,
		carWashRepository:   carWashRepository,
		userRepository:      userRepository,
		notificationService: notificationService,
	}
}

// CreateBooking for a selected time slot
func (bs *BookingService) CreateBooking(userID string, input models.Booking) (*models.Booking, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Convert userID to ObjectID
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Step 2: Validate BookingTime is not empty
	if input.BookingTime.IsZero() {
		return nil, errors.New("booking time is required")
	}

	//  Step 3: FETCH the Carwash and VALIDATE against its OpenHours
	carwash, err := bs.carWashRepository.GetCarwashByID(input.CarwashID)
	if err != nil {
		return nil, errors.New("carwash not found")
	}

	// Get the weekday from booking time (use full name to match open_hours keys)
	weekday := strings.ToLower(input.BookingTime.Weekday().String()) // "monday", "tuesday", etc

	logrus.Infof("Checking open hours for carwash %s on %s (derived from %v)", input.CarwashID, weekday, input.BookingTime)
	logrus.Infof("Carwash OpenHours keys: %v", carwash.OpenHours)

	timeRange, ok := carwash.OpenHours[weekday]
	if !ok {
		logrus.Warnf("Carwash %s is not open on %s. Available days: %v", input.CarwashID, weekday, carwash.OpenHours)
		return nil, errors.New("carwash is not open on this day")
	}

	// Convert time strings to time.Time
	layout := "15:04"
	bookingHour := input.BookingTime.Format(layout)

	bookingParsed, err := time.Parse(layout, bookingHour)
	if err != nil {
		return nil, errors.New("invalid booking time format")
	}

	startParsed, err := time.Parse(layout, timeRange.Start)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}

	endParsed, err := time.Parse(layout, timeRange.End)
	if err != nil {
		return nil, errors.New("invalid end time format")
	}

	// Compare booking time with open hours
	if bookingParsed.Before(startParsed) || bookingParsed.After(endParsed) {
		return nil, errors.New("booking time is outside of open hours")
	}

	// Step 4: Check for existing bookings on same date/time
	bookingsForDay, err := bs.bookingRepository.GetBookingsByDate(input.CarwashID, input.BookingTime)
	if err != nil {
		return nil, errors.New("could not fetch bookings for that time")
	}

	maxCars := carwash.MaxCarsPerSlot
	if maxCars <= 0 {
		maxCars = 1
	}

	currentCarsInSlot := 0
	inputTimeStr := input.BookingTime.UTC().Truncate(time.Minute).Format("2006-01-02 15:04")
	logrus.Infof("[CreateDebug] User: %s, Slot: %s, Max: %d", userID, inputTimeStr, maxCars)

	for _, b := range bookingsForDay {
		bTimeStr := b.BookingTime.UTC().Truncate(time.Minute).Format("2006-01-02 15:04")

		// If same user already has a pending/confirmed booking for this slot, block them
		if b.UserID.Hex() == userID && bTimeStr == inputTimeStr && b.Status != "cancelled" && b.Status != "completed" {
			return nil, errors.New("you already have a booking for this time slot")
		}

		if b.Status != "confirmed" {
			continue
		}

		if bTimeStr == inputTimeStr {
			currentCarsInSlot++
		}
	}

	if currentCarsInSlot >= maxCars {
		return nil, errors.New("selected time slot is already fully booked")
	}

	// Step 5: Create new booking
	queueNumber := len(bookingsForDay) + 1
	newBooking := models.Booking{

		ID:           primitive.NewObjectID(),
		UserID:       ownerID,
		CarID:        input.CarID,
		CarwashID:    input.CarwashID,
		BookingTime:  input.BookingTime,
		BookingType:  input.BookingType,
		UserLocation: input.UserLocation,
		AddressNote:  input.AddressNote,
		Notes:        input.Notes,
		Status:       "pending",
		QueueNumber:  queueNumber,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Check distance if it's a home service
	if input.BookingType == "home_service" {
		if input.UserLocation == nil {
			return nil, errors.New("user location is required for home service")
		}

		userLat := input.UserLocation.Coordinates[1]
		userLng := input.UserLocation.Coordinates[0]

		carwashLat := carwash.Location.Coordinates[1]
		carwashLng := carwash.Location.Coordinates[0]

		distance := utils.CalculateDistance(userLat, userLng, carwashLat, carwashLng)

		if float64(distance) > float64(carwash.DeliveryRadiusKM) {
			return nil, errors.New("user is outside the delivery radius for this carwash")
		}
	}

	// Step 6: Save to database
	if err := bs.bookingRepository.CreateBooking(&newBooking); err != nil {
		return nil, err
	}

	// Step 7: Trigger notification (Owner Notification)
	// Alert the Business Owner about the new Pending Booking
	// We need the customer name.
	go func() {
		user, err := bs.userRepository.FindUserByID(ownerID)
		if err == nil && bs.notificationService != nil {
			// Find Business Owner ID from Carwash (Assuming carwash has OwnerID)
			// Wait, carwash model needs checking. models.Carwash usually has OwnerID.
			// Re-fetching carwash to be sure or using cached 'carwash' variable if safe (it is local)

			// Check if OwnerID exists (handle legacy data)
			if !carwash.OwnerID.IsZero() {
				bs.notificationService.SendNewBookingToBusiness(carwash.OwnerID, user.Name, "New Booking")
			} else {
				logrus.Warnf("Carwash %s has no OwnerID, cannot send notification", carwash.Name)
			}
		}
	}()

	return &newBooking, nil

}

func (bs *BookingService) GetBookingByID(bookingID string) (*models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

func (bs *BookingService) GetBookingsByUserID(userID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	bookings, err := bs.bookingRepository.GetBookingsByUserID(objID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (bs *BookingService) GetBookingsByCarwashID(carwashID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	bookings, err := bs.bookingRepository.GetBookingsByCarwashID(objID)
	if err != nil {
		return nil, err
	}

	return bs.enrichBookingsWithCustomerDetails(bookings)
}

func (bs *BookingService) UpdateBookingStatus(bookingID string, newStatus string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// Fetch booking (needed for notifications)
	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return errors.New("booking not found")
	}

	if err := bs.bookingRepository.UpdateBookingStatus(objID, newStatus); err != nil {
		return err
	}

	// Trigger Notifications (Async)
	go func() {
		if bs.notificationService == nil {
			return
		}

		// Fetch carwash name
		var carwashName string
		cw, err := bs.carWashRepository.GetCarwashByID(booking.CarwashID)
		if err == nil {
			carwashName = cw.Name
		} else {
			carwashName = "The Carwash"
		}

		if newStatus == "confirmed" {
			// In-App + Email (Hybrid Strategy)
			bs.notificationService.SendBookingAccepted(booking, carwashName)
		} else if newStatus == "cancelled" {
			bs.notificationService.SendBookingRejected(booking, "Cancelled by business")
		} else if newStatus == "completed" {
			// In-App Only (Hybrid Strategy)
			title := "Wash Completed"
			message := fmt.Sprintf("Your service at %s is marked as completed. Please rate your experience!", carwashName)
			// false = No Email
			bs.notificationService.CreateNotification(booking.UserID, title, message, "booking", false)
		}
	}()

	return nil
}

func (bs *BookingService) CancelBooking(bookingID string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// 1. Fetch booking to check time
	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return errors.New("booking not found")
	}

	// 2. Check if already cancelled or completed
	if booking.Status == "cancelled" {
		return errors.New("booking is already cancelled")
	}
	if booking.Status == "completed" {
		return errors.New("cannot cancel a completed booking")
	}

	// 3. Enforce 24-hour cancellation policy
	// If booking time is within 24 hours from now, deny cancellation
	timeUntilBooking := booking.BookingTime.Sub(time.Now())
	if timeUntilBooking < 24*time.Hour && timeUntilBooking > 0 {
		return errors.New("cancellation not allowed within 24 hours of appointment")
	}

	return bs.bookingRepository.CancelBooking(objID)
}

func (bs *BookingService) GetBookingsByDate(carwashID string, date time.Time) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return bs.bookingRepository.GetBookingsByDate(objID, date)
}

func (bs *BookingService) UpdateBooking(userID, bookingID string, updates map[string]interface{}) error {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	bookingObjID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// Optional: check ownership here if needed

	// Add updatedAt
	updates["updated_at"] = time.Now()

	return bs.bookingRepository.UpdateBooking(bookingObjID, bson.M(updates))
}

// Get Bookings with filter
func (bs *BookingService) GetBookingsByCarwashWithFilters(carwashID string, status string, from, to string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	var fromDate, toDate time.Time
	if from != "" && to != "" {
		fromDate, err = time.Parse("2006-01-02", from)
		if err != nil {
			return nil, errors.New("invalid from date format")
		}
		toDate, err = time.Parse("2006-01-02", to)
		if err != nil {
			return nil, errors.New("invalid to date format")
		}
		// Adjust toDate to include the full day
		toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, toDate.Location())
	}

	bookings, err := bs.bookingRepository.GetBookingsByCarwashWithFilters(objID, status, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	return bs.enrichBookingsWithCustomerDetails(bookings)
}

// Helper to enrich bookings with customer details
func (bs *BookingService) enrichBookingsWithCustomerDetails(bookings []models.Booking) ([]models.Booking, error) {
	if len(bookings) == 0 {
		return bookings, nil
	}

	// Collect User IDs
	userIDs := make([]primitive.ObjectID, 0)
	seen := make(map[primitive.ObjectID]bool)
	for _, b := range bookings {
		if !seen[b.UserID] {
			userIDs = append(userIDs, b.UserID)
			seen[b.UserID] = true
		}
	}

	// Fetch Users
	users, err := bs.userRepository.GetUsersByIDs(userIDs)
	if err != nil {
		// Log error but return bookings without names rather than failing?
		// Or fail? Let's fail for now or log.
		// For now, return error.
		return nil, err
	}

	// Map Users
	userMap := make(map[primitive.ObjectID]models.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Enrich Bookings
	for i := range bookings {
		if user, ok := userMap[bookings[i].UserID]; ok {
			bookings[i].CustomerName = user.Name
			bookings[i].CustomerPhoto = user.ProfilePhoto
		}
	}

	return bookings, nil
}
