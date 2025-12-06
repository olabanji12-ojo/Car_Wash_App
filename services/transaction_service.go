package services

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//  CreatePayment
func CreatePayment(ownerID string, input models.Payment) (*models.Payment, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	UserID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}


	newPayment := models.Payment{
	ID:         primitive.NewObjectID(),
	OrderID:    input.OrderID,
	UserID:     UserID,
	CarwashID:  input.CarwashID,
	Amount:     input.Amount,
	Status:     "paid",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}


	if input.Status == "" {
		input.Status = "paid"
	}

	err = repositories.CreatePayment(&newPayment)
	
	if err != nil {
		return nil, err
	}

	return &newPayment, nil
}


//  GetPaymentByOrderID
func GetPaymentByOrderID(orderID string) (*models.Payment, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID")
	}

	return repositories.GetPaymentByOrderID(objID)
}


//  GetPaymentsByUserID
func GetPaymentsByUserID(userID string) ([]models.Payment, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetPaymentsByUserID(objID)
}


//  GetPaymentsByCarwashID
func GetPaymentsByCarwashID(carwashID string) ([]models.Payment, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return repositories.GetPaymentsByCarwashID(objID)
}



//  CalculateEarningsByCarwash
func CalculateEarningsByCarwash(carwashID string) (float64, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return 0, errors.New("invalid carwash ID")
	}

	return repositories.CalculateEarningsByCarwash(objID)
}



