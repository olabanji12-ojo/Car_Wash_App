package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/sirupsen/logrus"
	"fmt"

)

//  1. Create a new carwash
func CreateCarwash(carwash models.Carwash) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := database.CarwashCollection.InsertOne(ctx, carwash)
	if err != nil {
		return nil, err
	}
	return result, nil
}


//  2. Get a carwash by ID
func GetCarwashByID(objID primitive.ObjectID) (*models.Carwash, error) {
	
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var carwash models.Carwash
	err := database.CarwashCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&carwash)
	if err != nil {
		return nil, err
	}

	return &carwash, nil
}

//  3. Get all active carwashes (for customers to browse)
func GetActiveCarwashes() ([]models.Carwash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var carwashes []models.Carwash
	for cursor.Next(ctx) {
		var cw models.Carwash
		if err := cursor.Decode(&cw); err != nil {
			logrus.Info("Error decoding carwash:", err)
		}
		carwashes = append(carwashes, cw)
	}
	fmt.Println("Carwashes found:", len(carwashes))
	return carwashes, nil
	
}

//  4. Update a carwash
func UpdateCarwash(id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid carwash ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	return err
}

//  5. Toggle carwash online status
func SetCarwashStatus(id string, isActive bool) error {
	return UpdateCarwash(id, bson.M{"is_active": isActive})
}

//  6. Get all carwashes by business owner ID
func GetCarwashesByOwnerID(ownerID string) ([]models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, bson.M{"owner_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Carwash
	for cursor.Next(ctx) {
		var carwash models.Carwash
		if err := cursor.Decode(&carwash); err == nil {
			results = append(results, carwash)
		}
	}
	return results, nil
}

//  7. Filter carwashes (optional advanced search)
func GetCarwashesByFilter(filter bson.M) ([]models.Carwash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var carwashes []models.Carwash
	for cursor.Next(ctx) {
		var c models.Carwash
		if err := cursor.Decode(&c); err == nil {
			carwashes = append(carwashes, c)
		}
	}
	return carwashes, nil
}

//  8. Update queue count
func UpdateQueueCount(id string, count int) error {
	return UpdateCarwash(id, bson.M{"queue_count": count})
}