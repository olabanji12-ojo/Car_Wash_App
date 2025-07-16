package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//  1. CreatePayment
func CreatePayment(payment *models.Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.PaymentCollection.InsertOne(ctx, payment)
	return err
}

//  2. GetPaymentByOrderID
func GetPaymentByOrderID(orderID primitive.ObjectID) (*models.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var payment models.Payment
	err := database.PaymentCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	return &payment, nil
}

//  3. GetPaymentsByUserID
func GetPaymentsByUserID(userID primitive.ObjectID) ([]models.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.PaymentCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	for cursor.Next(ctx) {
		var payment models.Payment
		if err := cursor.Decode(&payment); err == nil {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

//  4. GetPaymentsByCarwashID
func GetPaymentsByCarwashID(carwashID primitive.ObjectID) ([]models.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.PaymentCollection.Find(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	for cursor.Next(ctx) {
		var payment models.Payment
		if err := cursor.Decode(&payment); err == nil {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

//  5. CalculateEarningsByCarwash
func CalculateEarningsByCarwash(carwashID primitive.ObjectID) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.M{"$match": bson.M{
		"carwash_id":     carwashID,
		"payment_status": "paid",
	}}

	groupStage := bson.M{"$group": bson.M{
		"_id":   nil,
		"total": bson.M{"$sum": "$amount"},
	}}

	cursor, err := database.PaymentCollection.Aggregate(ctx, []bson.M{matchStage, groupStage})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0, nil // No earnings yet
	}

	total := result[0]["total"].(float64)
	return total, nil
}

//  6. GetPaymentByReference
func GetPaymentByReference(reference string) (*models.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var payment models.Payment
	err := database.PaymentCollection.FindOne(ctx, bson.M{"reference": reference}).Decode(&payment)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	return &payment, nil
}



