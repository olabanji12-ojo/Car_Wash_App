package repositories

import (
	"context"
	"errors"
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

// FindWorkersByCarwashID gets all workers for a carwash
func (wr *WorkerRepository) FindWorkersByCarwashID(carwashID primitive.ObjectID) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"role":       "worker",
		"carwash_id": carwashID,
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
		"role":          "worker",
		"status":        "active", // Account is active
		"carwash_id":    businessID,
		"worker_status": "online",           // Worker is online
		"active_orders": bson.M{"$size": 0}, // No active orders
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
			"worker_status": workStatus,
			"last_seen":     time.Now(),
			"updated_at":    time.Now(),
		},
	}

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}

// AssignWorkerToBooking assigns a worker to a booking
func (wr *WorkerRepository) AssignWorkerToBooking(bookingID primitive.ObjectID, workerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bookingID}
	update := bson.M{
		"$set": bson.M{
			"worker_id":  workerID,
			"updated_at": time.Now(),
		},
	}

	result, err := wr.db.Collection("bookings").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("booking not found in database")
	}
	return nil
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
			"worker_status": "busy",
			"last_seen":     time.Now(),
			"updated_at":    time.Now(),
		},
		"$push": bson.M{
			"active_orders": orderID,
		},
	}

	result, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("worker not found in database during status update")
	}
	return nil
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
			"worker_status": "online",
			"last_seen":     time.Now(),
			"updated_at":    time.Now(),
		},
		"$pull": bson.M{
			"active_orders": orderID,
		},
	}

	_, err := wr.db.Collection("users").UpdateOne(ctx, filter, update)
	return err
}
