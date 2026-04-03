package services

import (
	"time"

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

// GetEarningsSummary returns the daily, weekly, and monthly earnings for a shop
func (ts *TransactionService) GetEarningsSummary(carwashID string) (map[string]float64, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]float64)

	// 1. Today
	today := time.Now().Truncate(24 * time.Hour)
	daily, _ := repositories.GetEarningsByPeriod(objID, today)
	summary["daily"] = daily

	// 2. This Week
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	weekly, _ := repositories.GetEarningsByPeriod(objID, oneWeekAgo)
	summary["weekly"] = weekly

	// 3. This Month
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	monthly, _ := repositories.GetEarningsByPeriod(objID, oneMonthAgo)
	summary["monthly"] = monthly

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
