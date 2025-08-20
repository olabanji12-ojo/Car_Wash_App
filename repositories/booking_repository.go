package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository struct { 
	db *mongo.Database 
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{db: db}
}


// 1. CreateBooking
func (br *BookingRepository) CreateBooking(booking *models.Booking) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.InsertOne(ctx, booking)
	if err != nil {
		logrus.Error("Failed to create booking: ", err)
		return err
	}
	return nil
}

// 2. GetBookingByID
func (br *BookingRepository) GetBookingByID(id primitive.ObjectID) (*models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var booking models.Booking
	err := database.BookingCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		logrus.Error("Booking not found: ", err)
		return nil, errors.New("booking not found")
	}
	return &booking, nil
}

// 3. GetBookingsByUserID
func (br *BookingRepository) GetBookingsByUserID(userID primitive.ObjectID) ([]models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.BookingCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		logrus.Error("Failed to fetch bookings by user: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		} else {
			logrus.Warn("Error decoding booking: ", err)
		}
	}
	return bookings, nil
}

// 4. GetBookingsByCarwashID
func (br *BookingRepository) GetBookingsByCarwashID(carwashID primitive.ObjectID) ([]models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.BookingCollection.Find(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		logrus.Error("Failed to fetch bookings by carwash: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		} else {
			logrus.Warn("Error decoding booking: ", err)
		}
	}
	return bookings, nil
}

// 5. UpdateBookingStatus
func (br *BookingRepository) UpdateBookingStatus(id primitive.ObjectID, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": newStatus, "updated_at": time.Now()}},
	)
	if err != nil {
		logrus.Error("Failed to update booking status: ", err)
		return err
	}
	return nil
}

// UpdateBooking (general update)
func (br *BookingRepository) UpdateBooking(bookingID primitive.ObjectID, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bookingID}
	update := bson.M{"$set": updates}

	_, err := database.BookingCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		logrus.Error("Failed to update booking: ", err)
		return err
	}
	return nil
}

// 7. CancelBooking (soft delete)
func (br *BookingRepository) CancelBooking(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": "cancelled", "updated_at": time.Now()}},
	)
	if err != nil {
		logrus.Error("Failed to cancel booking: ", err)
		return err
	}
	return nil
}

// 8. GetBookingsByDate
func (br *BookingRepository) GetBookingsByDate(carwashID primitive.ObjectID, date time.Time) ([]models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	filter := bson.M{
		"carwash_id":   carwashID,
		"booking_time": bson.M{"$gte": start, "$lt": end},
	}

	cursor, err := database.BookingCollection.Find(ctx, filter)
	if err != nil {
		logrus.Error("Failed to fetch bookings by date: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		} else {
			logrus.Warn("Error decoding booking: ", err)
		}
	}
	return bookings, nil
}

