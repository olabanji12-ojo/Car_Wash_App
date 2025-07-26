package services

import (
	
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"

)


//  CreateBooking for a selected time slot
func CreateBooking(userID string, input models.Booking) (*models.Booking, error) {
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
	carwash, err := repositories.GetCarwashByID(input.CarwashID)
	if err != nil {
		return nil, errors.New("carwash not found")
	}

	// Get the weekday from booking time
	weekday := strings.ToLower(input.BookingTime.Weekday().String()[:3]) // "mon", "tue", etc
	timeRange, ok := carwash.OpenHours[weekday]
	if !ok {
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

	// Step 4: Check for existing bookings on same date
	bookingsForDay, err := repositories.GetBookingsByDate(input.CarwashID, input.BookingTime)
	if err != nil {
		return nil, errors.New("could not fetch bookings for that time")
	}
	for _, b := range bookingsForDay {
		if b.BookingTime.Equal(input.BookingTime) {
			return nil, errors.New("selected time slot is already taken")
		}
	}

	// Step 5: Create new booking
	queueNumber := len(bookingsForDay) + 1
	newBooking := models.Booking{
        
         
		ID:           primitive.NewObjectID(),
		UserID:       ownerID,
		CarID:        input.CarID,
		CarwashID:    input.CarwashID,
		ServiceIDs:   input.ServiceIDs,
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
	if err := repositories.CreateBooking(&newBooking); err != nil {
		return nil, err
	}

	// Step 7: Trigger notification (like Django signal)
	// This runs asynchronously so it doesn't block the booking creation
	go func() {
		// Import notification service when we add it
		// NotificationSvc.SendBookingConfirmation(&newBooking)
	}()

	return &newBooking, nil

}



func GetBookingByID(bookingID string) (*models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	booking, err := repositories.GetBookingByID(objID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}


func GetBookingsByUserID(userID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	bookings, err := repositories.GetBookingsByUserID(objID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}


func GetBookingsByCarwashID(carwashID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	bookings, err := repositories.GetBookingsByCarwashID(objID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}


func UpdateBookingStatus(bookingID string, newStatus string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	return repositories.UpdateBookingStatus(objID, newStatus)
}


func CancelBooking(bookingID string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	return repositories.CancelBooking(objID)
}

func GetBookingsByDate(carwashID string, date time.Time) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return repositories.GetBookingsByDate(objID, date)
}


func UpdateBooking(userID, bookingID string, updates map[string]interface{}) error {
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

	return repositories.UpdateBooking(bookingObjID, bson.M(updates))
}



