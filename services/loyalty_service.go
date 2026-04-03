package services

import (
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoyaltyService struct {
	orderRepo repositories.OrderRepository
}

func NewLoyaltyService(orderRepo repositories.OrderRepository) *LoyaltyService {
	return &LoyaltyService{
		orderRepo: orderRepo,
	}
}

// CheckForRewardEligibility counts completed washes to see if the user gets the 6th free
func (ls *LoyaltyService) IsEligibleForFreeWash(userID string) (bool, int, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, 0, err
	}

	// 1. Fetch all completed orders for this user
	orders, err := ls.orderRepo.GetOrdersByUserID(userObjID)
	if err != nil {
		return false, 0, err
	}

	// 2. Filter for "completed" orders only
	completedCount := 0
	for _, order := range orders {
		if order.Status == "completed" {
			completedCount++
		}
	}

	// 3. Simple threshold logic: Every 6th wash is free (after 5 completed)
	// Example: 5 completed -> next is reward (free)
	moduloValue := (completedCount % 6)
	isEligible := (moduloValue == 5)

	return isEligible, completedCount, nil
}
