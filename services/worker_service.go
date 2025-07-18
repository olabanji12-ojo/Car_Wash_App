package services

import (
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ✅ Create new worker (called by business)
func CreateWorkerService(requester models.User, input models.User) error {
	if requester.Role != "business" {
		return errors.New("only business users can create workers")
	}

	input.ID = primitive.NewObjectID()
	input.Role = "worker"
	input.Status = "active"
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	input.WorkerData = &models.WorkerProfile{
		BusinessID: requester.ID.Hex(),
		JobRole:    input.WorkerData.JobRole,
	}

	return repositories.CreateUser(input)
}

// ✅ Get all workers under a business
func GetWorkersByBusinessID(businessID string) ([]*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, errors.New("invalid business ID format")
	}
	return repositories.FindWorkersByBusinessID(objID)
}

// ✅ Update a worker's status
func SetWorkerStatus(workerID string, status string) error {
	objID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID")
	}
	return repositories.UpdateWorkerStatus(objID, status)
}
