package services

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ✅ CreateOrderFromBooking
func CreateOrderFromBooking(bookingID string) (*models.Order, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Convert ID
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	// 2. Fetch the booking
	booking, err := repositories.GetBookingByID(objID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	// 3. Prevent duplicate order
	if booking.Status == "approved" || booking.Status == "completed" {
		return nil, errors.New("booking already approved or processed")
	}

	// 4. Build the order
	newOrder := models.Order{

		ID:            primitive.NewObjectID(),
		BookingID:     booking.ID,
		UserID:        booking.UserID,
		CarID:         booking.CarID,
		CarwashID:     booking.CarwashID,
		ServiceIDs:    booking.ServiceIDs,
		QueueNumber:   booking.QueueNumber,
		BookingType:   booking.BookingType,
		UserLocation:  booking.UserLocation,
		Status:        "active",
		TotalAmount:   0, // optional: calculate later
		PaymentStatus: "unpaid",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		
	}

	// 5. Save the order
	if err := repositories.CreateOrder(&newOrder); err != nil {
		logrus.Error("Failed to create order: ", err)
		return nil, err
	}

	// 6. Update booking status to approved
	if err := repositories.UpdateBookingStatus(booking.ID, "approved"); err != nil {
		logrus.Warn("Order created, but failed to update booking status")
	}

	return &newOrder, nil
}


func GetOrderByID(orderID string) (*models.Order, error) {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID")
	}

	return repositories.GetOrderByID(objID)
}



func GetOrdersByUser(userID string) ([]models.Order, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetOrdersByUserID(uid)
}

func GetOrdersByCarwash(carwashID string) ([]models.Order, error) {

	owner_id, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetOrdersByCarwashID(owner_id)
	

}



func UpdateOrderStatus(orderID string, newStatus string) error {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID")
	}

	return repositories.UpdateOrderStatus(objID, newStatus)
}


func AssignWorker(orderID string, workerID string) error {
	oid, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID")
	}

	wid, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID")
	}

	return repositories.AssignWorker(oid, wid)
}



