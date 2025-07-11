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

//  CreateService
func CreateService(userID string, input models.Service) (*models.Service, error) {

	
	carwash, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	
	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build the service object
	newService := models.Service{
		ID:          primitive.NewObjectID(),
		CarwashID:   carwash,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Duration:    input.Duration,
		IsAddon:     input.IsAddon,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repositories.CreateService(ctx, &newService); err != nil {
		logrus.Error("Failed to create service: ", err)
		return nil, err
	}

	return &newService, nil
}

//  GetServiceByID
func GetServiceByID(serviceID string) (*models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, errors.New("invalid service ID")
	}

	return repositories.GetServiceByID(ctx, objID)
}

//  GetServicesByCarwashID
func GetServicesByCarwashID(carwashID string) ([]models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return repositories.GetServicesByCarwashID(ctx, objID)
}

//  UpdateService
func UpdateService(userID string,  serviceID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return errors.New("invalid service ID")
	}

	// Automatically update timestamp
	updates["updated_at"] = time.Now()

	return repositories.UpdateService(ctx, objID, bson.M(updates))
}

//  DeleteService (soft delete)
func DeleteService(userID string, serviceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return errors.New("invalid service ID")
	}

	return repositories.DeleteService(ctx, objID)
}

