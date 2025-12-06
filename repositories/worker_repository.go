package repositories

import (
	"context"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// WorkerRepository handles database operations for workers
type WorkerRepository struct {
	db *mongo.Database
}

// NewWorkerRepository creates a new WorkerRepository instance
func NewWorkerRepository(db *mongo.Database) *WorkerRepository {
	return &WorkerRepository{db: db}
}

// FindWorkersByBusinessID gets all workers for a business
func (wr *WorkerRepository) FindWorkersByBusinessID(businessID primitive.ObjectID) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"role":                    "worker",
		"worker_data.business_id": businessID.Hex(),
	}

	cursor, err := wr.db.Collection("users").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var workers []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		workers = append(workers, &user)
	}
	
	logrus.Infof("Found %d workers for business", len(workers))
	return workers, nil
}

// FindWorkerByID gets a specific worker by ID
func (wr *WorkerRepository) FindWorkerByID(workerID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":  workerID,
		"role": "worker",
	}

	var worker models.User
	err := wr.db.Collection("users").FindOne(ctx, filter).Decode(&worker)
	if err != nil {
		return nil, err
	}

	return &worker, nil
}

// FindAvailableWorkersByBusinessID gets available workers for assignment (online and not busy)
func (wr *WorkerRepository) FindAvailableWorkersByBusinessID(businessID primitive.ObjectID) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"role":                      "worker",
		"status":                    "active",               // Account is active
		"worker_data.business_id":   businessID.Hex(),
		"worker_data.status":        "online",              // Worker is online
		"worker_data.active_orders": bson.M{"$size": 0},    // No active orders
	}

	cursor, err := wr.db.Collection("users").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var workers []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		workers = append(workers, &user)
	}
	
	logrus.Infof("Found %d available workers for business", len(workers))
	return workers, nil
}

// UpdateWorkerStatus updates a worker's account status
func (wr *WorkerRepository) UpdateWorkerStatus(workerID primitive.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}

// UpdateWorkerWorkStatus updates a worker's work status (online, offline, busy, on_break)
func (wr *WorkerRepository) UpdateWorkerWorkStatus(workerID primitive.ObjectID, workStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"worker_data.status":    workStatus,
			"worker_data.last_seen": time.Now(),
			"updated_at":            time.Now(),
		},
	}

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}

// AssignWorkerToOrder assigns a worker to an order
func (wr *WorkerRepository) AssignWorkerToOrder(orderID primitive.ObjectID, workerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"worker_id":  workerID,
			"updated_at": time.Now(),
		},
	}

	_, err := wr.db.Collection("orders").UpdateOne(ctx, filter, update)
	return err
}

// AddActiveOrderToWorker adds an order to worker's active orders and sets status to busy
func (wr *WorkerRepository) AddActiveOrderToWorker(workerID primitive.ObjectID, orderID primitive.ObjectID) error {
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

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}

// RemoveWorkerFromOrder removes a worker from an order (rollback function)
func (wr *WorkerRepository) RemoveWorkerFromOrder(orderID primitive.ObjectID) error {
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

	_, err := wr.db.Collection("orders").UpdateOne(ctx, filter, update)
	return err
}

// RemoveActiveOrderFromWorker removes an order from worker's active orders and sets status back to online
func (wr *WorkerRepository) RemoveActiveOrderFromWorker(workerID primitive.ObjectID, orderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": workerID}
	update := bson.M{
		"$set": bson.M{
			"worker_data.status":    "online",
			"worker_data.last_seen": time.Now(),
			"updated_at":            time.Now(),
		},
		"$pull": bson.M{
			"worker_data.active_orders": orderID,
		},
	}

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}