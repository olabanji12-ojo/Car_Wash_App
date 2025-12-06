package services

import (
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WorkerService handles business logic for worker operations
type WorkerService struct {
	userRepo   *repositories.UserRepository
	workerRepo *repositories.WorkerRepository
}

// NewWorkerService creates a new WorkerService instance
func NewWorkerService(userRepo *repositories.UserRepository, workerRepo *repositories.WorkerRepository) *WorkerService {
	return &WorkerService{
		userRepo:   userRepo,
		workerRepo: workerRepo,
	}
}

// CreateWorker creates a new worker (called by business)
func (ws *WorkerService) CreateWorker(requester models.User, input models.User) error {
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

	return ws.userRepo.CreateUser(input)
}

// GetWorkersByBusinessID gets all workers under a business
func (ws *WorkerService) GetWorkersByBusinessID(businessID string) ([]*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, errors.New("invalid business ID format")
	}
	return ws.workerRepo.FindWorkersByBusinessID(objID)
}

// SetWorkerStatus updates a worker's status
func (ws *WorkerService) SetWorkerStatus(workerID string, status string) error {
	objID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID")
	}
	return ws.workerRepo.UpdateWorkerStatus(objID, status)
}

// GetAvailableWorkersForAssignment gets available workers for assignment (online and not busy)
func (ws *WorkerService) GetAvailableWorkersForAssignment(businessID string) ([]*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, errors.New("invalid business ID format")
	}
	return ws.workerRepo.FindAvailableWorkersByBusinessID(objID)
}

// SetWorkerWorkStatus updates a worker's work status (online, offline, busy, on_break)
func (ws *WorkerService) SetWorkerWorkStatus(workerID string, workStatus string) error {
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

	return ws.workerRepo.UpdateWorkerWorkStatus(objID, workStatus)
}

// AssignWorkerToOrder assigns worker to order (manual assignment by business)
func (ws *WorkerService) AssignWorkerToOrder(workerID string, orderID string) error {
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
	worker, err := ws.workerRepo.FindWorkerByID(workerObjID)
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
	err = ws.workerRepo.AssignWorkerToOrder(orderObjID, workerObjID)
	if err != nil {
		return errors.New("failed to assign worker to order: " + err.Error())
	}

	// 4. Update worker status and active orders
	err = ws.workerRepo.AddActiveOrderToWorker(workerObjID, orderObjID)
	if err != nil {
		// Rollback: Remove worker from order if worker update fails
		ws.workerRepo.RemoveWorkerFromOrder(orderObjID)
		return errors.New("failed to update worker status: " + err.Error())
	}

	return nil
}

// RemoveWorkerFromOrder removes worker from order (unassign worker)
func (ws *WorkerService) RemoveWorkerFromOrder(workerID string, orderID string) error {
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
	err = ws.workerRepo.RemoveWorkerFromOrder(orderObjID)
	if err != nil {
		return errors.New("failed to remove worker from order: " + err.Error())
	}

	// 3. Free up the worker (remove order from active orders and set status to online)
	err = ws.workerRepo.RemoveActiveOrderFromWorker(workerObjID, orderObjID)
	if err != nil {
		return errors.New("failed to free up worker: " + err.Error())
	}

	return nil
}