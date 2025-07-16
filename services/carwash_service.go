package services

import (
	// "context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateCarwash(ownerID string, input models.Carwash) (*models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}
	input.ID = primitive.NewObjectID()
	input.OwnerID = objID
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()
	input.QueueCount = 0
	input.IsActive = true

	_, err = repositories.CreateCarwash(input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func GetCarwashByID(id string) (*models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid carwash ID format")
	}

	return repositories.GetCarwashByID(objID)
}

func GetAllActiveCarwashes() ([]models.Carwash, error) {
	return repositories.GetActiveCarwashes()
}

func UpdateCarwash(id string, updateData map[string]interface{}) error {
	updateData["updated_at"] = time.Now()
	return repositories.UpdateCarwash(id, updateData)
}

func SetCarwashStatus(id string, isActive bool) error {

	return repositories.SetCarwashStatus(id, isActive)

}

func GetCarwashesByOwnerID(ownerID string) ([]models.Carwash, error) {

	return repositories.GetCarwashesByOwnerID(ownerID)

}

func SearchCarwashes(filter bson.M) ([]models.Carwash, error) {

	return repositories.GetCarwashesByFilter(filter)

}

func UpdateQueueCount(carwashID string, count int) error {

	return repositories.UpdateQueueCount(carwashID, count)
	
}