package repositories

import (
    
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sirupsen/logrus"
    
)

//  1. CreateBooking
func CreateBooking(booking *models.Booking) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.InsertOne(ctx, booking)
	return err
}

//  2. GetBookingByID
func GetBookingByID(id primitive.ObjectID) (*models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var booking models.Booking
	err := database.BookingCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		return nil, errors.New("booking not found")
	}
	return &booking, nil
}

//  3. GetBookingsByUserID
func GetBookingsByUserID(userID primitive.ObjectID) ([]models.Booking, error) {
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.BookingCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		}
	}
	return bookings, nil
}

//  4. GetBookingsByCarwashID
func GetBookingsByCarwashID(carwashID primitive.ObjectID) ([]models.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.BookingCollection.Find(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		}
	}
	return bookings, nil
}

//  5. UpdateBookingStatus
func UpdateBookingStatus(id primitive.ObjectID, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": newStatus, "updated_at": time.Now()}},
	)
	return err
}

func UpdateBooking(bookingID primitive.ObjectID, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bookingID}
	update := bson.M{"$set": updates}

	_, err := database.BookingCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		logrus.Error("Failed to update car: ", err)
		return err
	}
	return nil
}


//  6. GetLatestBookingTime
// func GetLatestBookingTime(carwashID primitive.ObjectID) (*time.Time, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	opts := bson.D{{"$sort", bson.D{{"booking_time", -1}}}}
// 	filter := bson.M{"carwash_id": carwashID}

// 	var booking models.Booking
// 	err := database.BookingCollection.FindOne(ctx, filter, opts).Decode(&booking)
// 	if err != nil {
// 		return nil, errors.New("no bookings found")
// 	}
// 	return &booking.BookingTime, nil
// }


//  7. CancelBooking (soft delete)
func CancelBooking(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookingCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": "cancelled", "updated_at": time.Now()}},
	)
	return err
}


//  8. GetBookingsByDate
func GetBookingsByDate(carwashID primitive.ObjectID, date time.Time) ([]models.Booking, error) {
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
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
		}
	}
	return bookings, nil
}



