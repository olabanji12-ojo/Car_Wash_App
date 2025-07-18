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
		"role": "worker",
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

// 🔄 Update a worker's status
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
