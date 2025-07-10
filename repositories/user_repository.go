package repositories

import (
	
	"context"
	// "errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"


)


// CreateUser inserts a new user into the MongoDB users collection
func CreateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.UserCollection.InsertOne(ctx, user)
	if err != nil {
		logrus.Error("Error inserting user: ", err)
		return err
	}
	logrus.Info("User inserted successfully")
	return nil
}


// FindUserByEmail searches for a user by their email
func FindUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := database.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		logrus.Warn("User not found with email: ", email)
		return nil, err
	}
	return &user, nil
}


// FindUserByID fetches a user by MongoDB ObjectID
func FindUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := database.UserCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}


// ✅ New: Update user profile fields
func UpdateUserByID(userID primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.UserCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},       // filter
		bson.M{"$set": update},      // update operation
	)

	return err
}


// FindWorkerByBusinessID


func FindWorkersByBusinessID(businessID primitive.ObjectID) ([]*models.User, error) {
	filter := bson.M{
		"role": "worker",
		"worker_data.business_id": businessID,
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

	return workers, nil
}





