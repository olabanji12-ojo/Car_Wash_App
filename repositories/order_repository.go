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


// THE REPOSITORY RELATES AND INTERACTS DIRECTLY WITH THE DATABASE 


// 1. CreateOrder - insert new order

func CreateOrder(order *models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.OrderCollection.InsertOne(ctx, order)
	if err != nil {
		logrus.Error("Failed to insert car: ", err)
		return err
	}

	return nil
}

// 2. GetOrderByID - fetch one order by ID
func GetOrderByID(orderID primitive.ObjectID) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var order models.Order
	err := database.OrderCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return nil, errors.New("order not found")
	}
	return &order, nil
}

// 3. GetOrdersByUserID - list of orders made by a user
func GetOrdersByUserID(userID primitive.ObjectID) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.OrderCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err == nil {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

// 4. GetOrdersByCarwashID - orders received by a carwash
func GetOrdersByCarwashID(carwashID primitive.ObjectID) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.OrderCollection.Find(ctx, bson.M{"carwash_id": carwashID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err == nil {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

// 5. UpdateOrderStatus - mark order as in_progress, completed, etc.
func UpdateOrderStatus(orderID primitive.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.OrderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

// 6. AssignWorker - attach a worker to this order
func AssignWorker(orderID, workerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.OrderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{
			"$set": bson.M{
				"worker_id":  workerID,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}
