package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ✅ CreateService inserts a new service
func CreateService(ctx context.Context, service *models.Service) error {
	_, err := database.ServiceCollection.InsertOne(ctx, service)
	return err
}

// ✅ GetServiceByID fetches a single service by ID
func GetServiceByID(ctx context.Context, serviceID primitive.ObjectID) (*models.Service, error) {
	var service models.Service
	err := database.ServiceCollection.FindOne(ctx, bson.M{"_id": serviceID}).Decode(&service)
	if err != nil {
		return nil, errors.New("service not found")
	}
	return &service, nil
}

// ✅ GetServicesByCarwashID fetches all active services for a carwash
func GetServicesByCarwashID(ctx context.Context, carwashID primitive.ObjectID) ([]models.Service, error) {
	cursor, err := database.ServiceCollection.Find(ctx, bson.M{
		"carwash_id": carwashID,
		"active":     true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []models.Service
	for cursor.Next(ctx) {
		var s models.Service
		if err := cursor.Decode(&s); err == nil {
			services = append(services, s)
		}
	}
	return services, nil
}

// ✅ UpdateService updates a service document
func UpdateService(ctx context.Context, serviceID primitive.ObjectID, updates bson.M) error {
	updates["updated_at"] = time.Now()

	_, err := database.ServiceCollection.UpdateOne(
		ctx,
		bson.M{"_id": serviceID},
		bson.M{"$set": updates},
	)
	return err
	
}

// ✅ DeleteService marks the service as inactive (soft delete)
func DeleteService(ctx context.Context, serviceID primitive.ObjectID) error {
	_, err := database.ServiceCollection.UpdateOne(
		ctx,
		bson.M{"_id": serviceID},
		bson.M{"$set": bson.M{
			"active":     false,
			"updated_at": time.Now(),
		}},
	)
	return err
}