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
	"go.mongodb.org/mongo-driver/mongo/options"
	"fmt"
	"encoding/json" 
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

	
	options := options.Find()
	options.SetSort(bson.D{{"created_at", -1}})

	cursor, err := database.BookingCollection.Find(ctx, bson.M{"user_id": userID}, options)
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

	// Calculate date range
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	// COMPREHENSIVE DEBUG LOGGING
	logrus.Info("🔍 REPOSITORY DEBUG:")
	logrus.Info("Input carwashID:", carwashID)
	logrus.Info("Input carwashID hex:", carwashID.Hex())
	logrus.Info("Input date:", date)
	logrus.Info("Calculated start time:", start)
	logrus.Info("Calculated end time:", end)
	logrus.Info("Date location:", date.Location())
	logrus.Info("Start time UTC:", start.UTC())
	logrus.Info("End time UTC:", end.UTC())

	filter := bson.M{
		"carwash_id": carwashID,
		"booking_time": bson.M{"$gte": start, "$lt": end},
	}

	logrus.Info("🔍 MONGODB FILTER:")
	logrus.Info("Filter:", filter)
	
	// Convert filter to JSON for better readability
	filterJSON, _ := json.Marshal(filter)
	logrus.Info("Filter as JSON:", string(filterJSON))

	// Check if there are any bookings for this carwash (regardless of date)
	logrus.Info("🔍 CHECKING TOTAL BOOKINGS FOR CARWASH:")
	totalCount, err := database.BookingCollection.CountDocuments(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		logrus.Error("❌ Error counting total bookings:", err)
	} else {
		logrus.Info("Total bookings for carwash:", totalCount)
	}

	// Check bookings without date filter
	logrus.Info("🔍 CHECKING ALL CARWASH BOOKINGS (no date filter):")
	allBookingsCursor, err := database.BookingCollection.Find(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		logrus.Error("❌ Error fetching all bookings:", err)
	} else {
		var allBookings []models.Booking
		if err := allBookingsCursor.All(ctx, &allBookings); err == nil {
			logrus.Info("All bookings count:", len(allBookings))
			for i, booking := range allBookings {
				logrus.Info(fmt.Sprintf("📋 ALL BOOKING %d:", i))
				logrus.Info("  ID:", booking.ID.Hex())
				logrus.Info("  CarwashID:", booking.CarwashID.Hex())
				logrus.Info("  BookingTime:", booking.BookingTime)
				logrus.Info("  BookingTime UTC:", booking.BookingTime.UTC())
				logrus.Info("  Status:", booking.Status)
			}
		}
		allBookingsCursor.Close(ctx)
	}

	// Now run the actual filtered query
	logrus.Info("🚀 EXECUTING FILTERED QUERY:")
	cursor, err := database.BookingCollection.Find(ctx, filter)
	if err != nil {
		logrus.Error("❌ Failed to fetch bookings by date:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	bookingCount := 0
	
	for cursor.Next(ctx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err == nil {
			bookings = append(bookings, booking)
			bookingCount++
			
			logrus.Info(fmt.Sprintf("✅ DECODED BOOKING %d:", bookingCount))
			logrus.Info("  ID:", booking.ID.Hex())
			logrus.Info("  CarwashID:", booking.CarwashID.Hex())
			logrus.Info("  BookingTime:", booking.BookingTime)
			logrus.Info("  BookingTime UTC:", booking.BookingTime.UTC())
			logrus.Info("  Status:", booking.Status)
			logrus.Info("  Date matches filter:", booking.BookingTime.After(start) && booking.BookingTime.Before(end))
		} else {
			logrus.Warn("⚠️ Error decoding booking:", err)
		}
	}

	logrus.Info("📊 FINAL RESULT:")
	logrus.Info("Total bookings returned:", len(bookings))
	
	if len(bookings) == 0 {
		logrus.Warn("⚠️ NO BOOKINGS FOUND - POSSIBLE ISSUES:")
		logrus.Warn("1. Check if carwash_id exists in database")
		logrus.Warn("2. Check if booking_time is stored correctly")
		logrus.Warn("3. Check timezone handling")
		logrus.Warn("4. Check if data exists for the specified date")
	}

	return bookings, nil
}


// GetBookingsByCarwashWithFilters retrieves bookings for a car wash filtered by status and date range
func (br *BookingRepository) GetBookingsByCarwashWithFilters(carwashID primitive.ObjectID, status string, from, to time.Time) ([]models.Booking, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{
        "carwash_id": carwashID,
    }
    if status != "" && status != "all" {
        filter["status"] = status
    }
    if !from.IsZero() && !to.IsZero() {
        filter["booking_time"] = bson.M{
            "$gte": from,
            "$lte": to,
        }
    }

    cursor, err := database.BookingCollection.Find(ctx, filter)
    if err != nil {
        logrus.Error("Failed to fetch bookings by carwash with filters: ", err)
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