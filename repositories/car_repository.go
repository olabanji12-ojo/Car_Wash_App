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

//  CreateCar inserts a new car document into the DB
func CreateCar(car *models.Car) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CarCollection.InsertOne(ctx, car)
	if err != nil {
		logrus.Error("Failed to insert car: ", err)
		return err
	}
	return nil
}

//  UnsetDefaultCarsForUser sets is_default = false for all of a user's cars
func UnsetDefaultCarsForUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"owner_id": userID, "is_default": true}
	update := bson.M{"$set": bson.M{"is_default": false}}

	_, err := database.CarCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		logrus.Error("Failed to unset default cars: ", err)
		return err
	}
	return nil
}

//  GetCarsByUserID retrieves all cars for a specific user
func GetCarsByUserID(userID primitive.ObjectID) ([]models.Car, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cars []models.Car
	filter := bson.M{"owner_id": userID}

	cursor, err := database.CarCollection.Find(ctx, filter)
	if err != nil {
		logrus.Error("Failed to fetch cars: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var car models.Car
		if err := cursor.Decode(&car); err != nil {
			logrus.Warn("Error decoding car: ", err)
			continue
		}
		cars = append(cars, car)
	}
	return cars, nil
}

//  UpdateCar modifies a car's data
func UpdateCar(carID primitive.ObjectID, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": carID}
	update := bson.M{"$set": updates}

	_, err := database.CarCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		logrus.Error("Failed to update car: ", err)
		return err
	}
	return nil
}

//  DeleteCar removes a car by ID
func DeleteCar(carID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CarCollection.DeleteOne(ctx, bson.M{"_id": carID})
	if err != nil {
		logrus.Error("Failed to delete car: ", err)
		return err
	}
	return nil
}

//  SetDefaultCar sets one car as default and unsets others
func SetDefaultCar(userID, carID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Unset all other defaults
	err := UnsetDefaultCarsForUser(userID)
	if err != nil {
		return err
	}

	// Step 2: Set this car as default
	filter := bson.M{"_id": carID, "owner_id": userID}
	update := bson.M{"$set": bson.M{"is_default": true}}

	_, err = database.CarCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		logrus.Error("Failed to set default car: ", err)
		return err
	}
	return nil
}

//  GetCarByID fetches a car by its ID
func GetCarByID(carID primitive.ObjectID) (*models.Car, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var car models.Car
	err := database.CarCollection.FindOne(ctx, bson.M{"_id": carID}).Decode(&car)
	if err != nil {
		logrus.Error("Car not found: ", err)
		return nil, err
	}
	return &car, nil
}
