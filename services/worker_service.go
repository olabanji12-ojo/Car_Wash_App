package services

import (

	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"


)

// ✅ Create new worker (called by business)
func CreateWorkerService(requester models.User, input models.User) error {

	if requester.AccountType != "carwash" || requester.Role != "business_owner" {
		return errors.New("only business owners can create workers")
	}

	input.ID = primitive.NewObjectID() 
	input.AccountType = utils.ACCOUNT_TYPE_CAR_WASH
	input.Role = utils.ROLE_WORKER
	input.Status = "active"
	input.WorkerStatus = "active"
	input.CarWashID = requester.CarWashID

	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

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

// ✅ Get available workers for assignment (online and not busy)
func GetAvailableWorkersForAssignment(businessID string) ([]*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, errors.New("invalid business ID format")
	}
	return repositories.FindAvailableWorkersByBusinessID(objID)
}

// ✅ Update a worker's work status (online, offline, busy, on_break)
func SetWorkerWorkStatus(workerID string, workStatus string) error {
	objID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID")
	}

	// Validate work status values
	validStatuses := []string{"online", "offline", "busy", "on_break"}
	isValid := false
	for _, validStatus := range validStatuses {
		if workStatus == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("invalid work status. Must be: online, offline, busy, or on_break")
	}

	return repositories.UpdateWorkerWorkStatus(objID, workStatus)
}


// ✅ Assign worker to order (manual assignment by business)
func AssignWorkerToOrder(workerID string, orderID string) error {
	// 1. Validate IDs
	workerObjID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID format")
	}

	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID format")
	}

	// 2. Check if worker is available (optional validation)
	worker, err := repositories.FindWorkerByID(workerObjID)
	if err != nil {
		return errors.New("worker not found")
	}

	if worker.WorkerStatus != "online" {
		return errors.New("worker is not available for assignment")
	}

	if len(worker.ActiveOrders) > 0 {
		return errors.New("worker is already busy with another order")
	}

	// 3. Update order with worker ID
	err = repositories.AssignWorkerToOrder(orderObjID, workerObjID)
	if err != nil {
		return errors.New("failed to assign worker to order: " + err.Error())
	}

	// 4. Update worker status and active orders
	err = repositories.AddActiveOrderToWorker(workerObjID, orderObjID)
	if err != nil {
		// Rollback: Remove worker from order if worker update fails
		repositories.RemoveWorkerFromOrder(orderObjID)
		return errors.New("failed to update worker status: " + err.Error())
	}

	return nil
}

// ✅ Remove worker from order (unassign worker)
func RemoveWorkerFromOrder(workerID string, orderID string) error {
	// 1. Validate IDs
	workerObjID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID format")
	}

	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID format")
	}

	// 2. Remove worker from order
	err = repositories.RemoveWorkerFromOrder(orderObjID)
	if err != nil {
		return errors.New("failed to remove worker from order: " + err.Error())
	}

	// 3. Free up the worker (remove order from active orders and set status to online)
	err = repositories.RemoveActiveOrderFromWorker(workerObjID, orderObjID)
	if err != nil {
		return errors.New("failed to free up worker: " + err.Error())
	}

	return nil
}
