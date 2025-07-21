package repositories

import (
	"context"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 🔎 Get all workers for a business
func FindWorkersByBusinessID(businessID primitive.ObjectID) ([]*models.User, error) {
	filter := bson.M{
		"role":                    "worker",
		"worker_data.business_id": businessID.Hex(),
	}

	cursor, err := database.UserCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.TODO())

	var workers []*models.User
	for cursor.Next(context.TODO()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		workers = append(workers, &user)
	}
	logrus.Infof("Found %d workers for business", len(workers))
	return workers, nil
}

// 🔎 Get a specific worker by ID
func FindWorkerByID(workerID primitive.ObjectID) (*models.User, error) {
	filter := bson.M{
		"_id":  workerID,
		"role": "worker",
	}

	var worker models.User
	err := database.UserCollection.FindOne(context.TODO(), filter).Decode(&worker)
	if err != nil {
		return nil, err
	}

	return &worker, nil
}

// � Get available workers for assignment (online and not busy)
func FindAvailableWorkersByBusinessID(businessID primitive.ObjectID) ([]*models.User, error) {
	filter := bson.M{
		"role":                      "worker",
		"status":                    "active", // Account is active
		"worker_data.business_id":   businessID.Hex(),
		"worker_data.status":        "online",           // Worker is online
		"worker_data.active_orders": bson.M{"$size": 0}, // No active orders
	}

	cursor, err := database.UserCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.TODO())

	var workers []*models.User
	for cursor.Next(context.TODO()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		workers = append(workers, &user)
	}
	logrus.Infof("Found %d available workers for business", len(workers))
	return workers, nil
}

// �🔄 Update a worker's status
func UpdateWorkerStatus(workerID primitive.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := database.UserCollection.UpdateOne(ctx, filter, update)
	return err
}

// 🔄 Update a worker's work status (online, offline, busy, on_break)
func UpdateWorkerWorkStatus(workerID primitive.ObjectID, workStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"worker_data.status":    workStatus, // Update WorkerProfile.WorkerStatus
			"worker_data.last_seen": time.Now(), // Update LastSeen timestamp
			"updated_at":            time.Now(), // Update main record timestamp
		},
	}

	_, err := database.UserCollection.UpdateOne(ctx, filter, update)
	return err
}

// 🔄 Assign worker to order
func AssignWorkerToOrder(orderID primitive.ObjectID, workerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"worker_id":  workerID,
			"updated_at": time.Now(),
		},
	}

	_, err := database.OrderCollection.UpdateOne(ctx, filter, update)
	return err
}

// 🔄 Add order to worker's active orders and set status to busy
func AddActiveOrderToWorker(workerID primitive.ObjectID, orderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"worker_data.status":    "busy",
			"worker_data.last_seen": time.Now(),
			"updated_at":            time.Now(),
		},
		"$push": bson.M{
			"worker_data.active_orders": orderID,
		},
	}

	_, err := database.UserCollection.UpdateOne(ctx, filter, update)
	return err
}

// 🔄 Remove worker from order (rollback function)
func RemoveWorkerFromOrder(orderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$unset": bson.M{
			"worker_id": "",
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := database.OrderCollection.UpdateOne(ctx, filter, update)
	return err
}

// 🔄 Remove order from worker's active orders and set status back to online
func RemoveActiveOrderFromWorker(workerID primitive.ObjectID, orderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"worker_data.status":    "online", // Set worker back to online
			"worker_data.last_seen": time.Now(),
			"updated_at":            time.Now(),
		},
		"$pull": bson.M{
			"worker_data.active_orders": orderID, // Remove order from active orders
		},
	}

	_, err := database.UserCollection.UpdateOne(ctx, filter, update)
	return err
}
