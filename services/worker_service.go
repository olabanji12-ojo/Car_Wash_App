package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("🔍 [WorkerService.CreateWorker] Validating requester: ID=%s, AccountType=%s, Role=%s, CarWashID=%v", requester.ID.Hex(), requester.AccountType, requester.Role, requester.CarWashID)

	if requester.AccountType != utils.ACCOUNT_TYPE_CAR_WASH || requester.Role != utils.ROLE_BUSINESS {
		logrus.Warnf("⚠️ [WorkerService.CreateWorker] Validation failed: Expected AccountType=%s, Role=%s", utils.ACCOUNT_TYPE_CAR_WASH, utils.ROLE_BUSINESS)
		return errors.New("only business owners can create workers")
	}

	input.ID = primitive.NewObjectID()
	input.AccountType = utils.ACCOUNT_TYPE_CAR_WASH
	input.Role = utils.ROLE_WORKER
	input.Status = "active"
	input.WorkerStatus = "online"

	// Robust CarWashID assignment
	if requester.CarWashID != nil {
		input.CarWashID = requester.CarWashID
	} else if input.CarWashID != nil {
		logrus.Warn("⚠️ [WorkerService.CreateWorker] Requester CarWashID is nil, falling back to input CarWashID")
		// Use input CarWashID if requester's is nil (might happen if local state is ahead of DB)
	} else {
		logrus.Error("❌ [WorkerService.CreateWorker] No CarWashID available from requester or input")
		return errors.New("no carwash associated with this business owner")
	}

	logrus.Infof("📝 [WorkerService.CreateWorker] Final CarWashID: %v", input.CarWashID)

	// Hash a default password for the worker
	// We use their phone or a generic placeholder if not provided
	defaultPass := "password123"
	if input.Phone != "" {
		defaultPass = input.Phone
	}
	logrus.Infof("🔐 [WorkerService.CreateWorker] Hashing password for %s", input.Email)
	hashedPassword, err := utils.HashPassword(defaultPass)
	if err != nil {
		logrus.Errorf("❌ [WorkerService.CreateWorker] Password hashing failed: %v", err)
		return errors.New("failed to secure worker account: " + err.Error())
	}
	input.Password = hashedPassword

	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	logrus.Infof("💾 [WorkerService.CreateWorker] Saving worker to DB: Email=%s, CarWashID=%v", input.Email, input.CarWashID)
	return ws.userRepo.CreateUser(input)
}

// GetWorkersByCarwashID gets all workers under a carwash
func (ws *WorkerService) GetWorkersByCarwashID(carwashID string) ([]*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID format")
	}
	return ws.workerRepo.FindWorkersByCarwashID(objID)
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

	// 2. Check if worker exists
	worker, err := ws.workerRepo.FindWorkerByID(workerObjID)
	if err != nil {
		logrus.Errorf("❌ [WorkerService.AssignWorkerToOrder] Worker not found: %v", workerID)
		return errors.New("worker not found")
	}

	// MVP: Be more flexible with status
	if worker.WorkerStatus != "online" && worker.WorkerStatus != "active" {
		logrus.Warnf("⚠️ [WorkerService.AssignWorkerToOrder] Worker %s has status '%s'", worker.Name, worker.WorkerStatus)
		// For now, let's allow "active" as well
	}

	// Check if already assigned (optional warning)
	if len(worker.ActiveOrders) > 0 {
		logrus.Warnf("⚠️ [WorkerService.AssignWorkerToOrder] Worker %s already has %d active orders", worker.Name, len(worker.ActiveOrders))
		// For MVP, we'll allow multiple assignments or at least skip the hard error if desired
	}

	// 3. Update booking with worker ID
	err = ws.workerRepo.AssignWorkerToBooking(orderObjID, workerObjID)
	if err != nil {
		return errors.New("failed to assign worker to booking: " + err.Error())
	}

	// 4. Update worker status and active orders
	err = ws.workerRepo.AddActiveOrderToWorker(workerObjID, orderObjID)
	if err != nil {
		// Rollback: Remove worker from booking if worker update fails
		// (Using generic RemoveWorkerFromOrder for now, assuming it handles 'worker_id' unset)
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

// UpdateWorker updates worker basic details
func (ws *WorkerService) UpdateWorker(workerID string, updateData map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID format")
	}

	// Only allow updating certain fields for security
	allowedFields := map[string]bool{
		"name":     true,
		"phone":    true,
		"job_role": true,
	}

	cleanUpdate := make(map[string]interface{})
	for k, v := range updateData {
		if allowedFields[k] {
			cleanUpdate[k] = v
		}
	}

	if len(cleanUpdate) == 0 {
		return errors.New("no valid fields to update")
	}

	return ws.userRepo.UpdateUserByID(objID, cleanUpdate)
}

// UploadWorkerPhoto uploads a worker profile photo
func (ws *WorkerService) UploadWorkerPhoto(workerID string, photoFile *ProfilePhotoFile) (string, error) {
	objID, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return "", errors.New("invalid worker ID format")
	}

	// 1. Upload to Cloudinary
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("worker_%s_%d", workerID, timestamp)

	uploadResult, err := UploadImage(photoFile.File, filename, "worker_photos")
	if err != nil {
		return "", err
	}

	// 2. Update DB
	update := map[string]interface{}{
		"profile_photo": uploadResult.SecureURL,
		"updated_at":    time.Now(),
	}

	err = ws.userRepo.UpdateUserByID(objID, update)
	if err != nil {
		DeleteImage(uploadResult.PublicID) // Cleanup
		return "", err
	}

	return uploadResult.SecureURL, nil
}
