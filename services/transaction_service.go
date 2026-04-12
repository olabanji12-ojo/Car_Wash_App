package services

import (
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionService struct {
	// No specific fields needed but good for consistency
}

func NewTransactionService() *TransactionService {
	return &TransactionService{}
}

// GetEarningsSummary returns the daily, weekly, and monthly earnings, plus totals for a shop
func (ts *TransactionService) GetEarningsSummary(carwashID string) (map[string]interface{}, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]interface{})

	// 1. Periods (Revenue)
	today := time.Now().Truncate(24 * time.Hour)
	daily, _ := repositories.GetEarningsByPeriod(objID, today)
	summary["daily"] = daily

	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	weekly, _ := repositories.GetEarningsByPeriod(objID, oneWeekAgo)
	summary["weekly"] = weekly

	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	monthly, _ := repositories.GetEarningsByPeriod(objID, oneMonthAgo)
	summary["monthly"] = monthly

	// 2. Totals (Revenue)
	totalRevenue, _ := repositories.CalculateEarningsByCarwash(objID)
	summary["total_revenue"] = totalRevenue

	// 3. Booking Metrics
	// Use our static-style lookup for now to match project pattern
	bookingRepo := repositories.NewBookingRepository(database.DB)
	bookings, _ := bookingRepo.GetBookingsByCarwashID(objID)
	
	totalBookings := len(bookings)
	pendingBookings := 0
	for _, b := range bookings {
		if b.Status == "pending" {
			pendingBookings++
		}
	}

	summary["total_bookings"] = totalBookings
	summary["pending_bookings"] = pendingBookings

	return summary, nil
}

// GetTransactionHistory returns the list of payments for a shop
func (ts *TransactionService) GetTransactionHistory(carwashID string) ([]models.Payment, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, err
	}
	return repositories.GetPaymentsByCarwashID(objID)
}

// GetTransactionHistoryByUserID returns the list of payments for a user
func (ts *TransactionService) GetTransactionHistoryByUserID(userID primitive.ObjectID) ([]models.Payment, error) {
	return repositories.GetPaymentsByUserID(userID)
}
