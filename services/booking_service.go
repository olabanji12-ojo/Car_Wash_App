package services

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)



//  CreateBooking for a selected time slot
func CreateBooking(userID string, input models.Booking) (*models.Booking, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//  Convert userID to ObjectID
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}


	// Validate BookingTime
	if input.BookingTime.IsZero() {
		return nil, errors.New("booking time is required")
	}

	
	//  Check if the selected slot is already taken for that carwash
	bookingsForDay, err := repositories.GetBookingsByDate(input.CarwashID, input.BookingTime)
	if err != nil {
		return nil, errors.New("could not fetch bookings for that time")
	}

	for _, b := range bookingsForDay {
		if b.BookingTime.Equal(input.BookingTime) {
			return nil, errors.New("selected time slot is already taken")
		}
	}

	// Calculate queue number for the day (e.g., 3rd person today = 3)
	queueNumber := len(bookingsForDay) + 1

	//  Build booking object
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

	//  Save to database
	if err := repositories.CreateBooking(&newBooking); err != nil {
		return nil, err
	}

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


