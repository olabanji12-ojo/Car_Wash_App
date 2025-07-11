package services

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//  CreateCar
func CreateCar(userID string, input models.Car) (*models.Car, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// If this is set as default, unset others first
	if input.IsDefault {
		if err := repositories.UnsetDefaultCarsForUser(ownerID); err != nil {
			return nil, err
		}
	}

	newCar := models.Car{
		ID:        primitive.NewObjectID(),
		OwnerID:   ownerID,
		Model:     input.Model,
		Plate:     input.Plate,
		Color:     input.Color,
		IsDefault: input.IsDefault,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repositories.CreateCar(&newCar); err != nil {
		logrus.Error("Failed to create car: ", err)
		return nil, err
	}

	
	return &newCar, nil
	
}

//  GetCarsByUserID
func GetCarsByUserID(userID string) ([]models.Car, error) {
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetCarsByUserID(ownerID)
}

//  UpdateCar
func UpdateCar(userID, carID string, updates map[string]interface{}) error {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	// Optional: check ownership here if needed

	// Add updatedAt
	updates["updated_at"] = time.Now()

	return repositories.UpdateCar(carObjID, bson.M(updates))
}

//  DeleteCar
func DeleteCar(userID, carID string) error {
	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	// Optional: confirm ownership logic before deleting
	return repositories.DeleteCar(carObjID)
}

//  SetDefaultCar
func SetDefaultCar(userID, carID string) error {
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	return repositories.SetDefaultCar(ownerID, carObjID)
}

//  GetCarByID (optional helper)
func GetCarByID(carID string) (*models.Car, error) {
	objID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return nil, errors.New("invalid car ID")
	}
	return repositories.GetCarByID(objID)
}



