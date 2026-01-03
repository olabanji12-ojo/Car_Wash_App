package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
)

type WorkerController struct {
	WorkerService *services.WorkerService
	UserService   *services.UserService
}

func NewWorkerController(workerService *services.WorkerService, userService *services.UserService) *WorkerController {
	return &WorkerController{
		WorkerService: workerService,
		UserService:   userService,
	}
}

// CreateWorkersForBusiness creates a new worker for a business
func (wc *WorkerController) CreateWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	// Get authenticated business user
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		logrus.Errorf("❌ [WorkerController] Auth context failed: %v", err)
		utils.Error(w, http.StatusUnauthorized, "Invalid authentication context")
		return
	}
	logrus.Infof("👤 [WorkerController] Requester: ID=%s, Role=%s, AccountType=%s", authCtx.UserID, authCtx.Role, authCtx.AccountType)
	requesterUserID := authCtx.UserID

	// Get requester (business owner) details
	requester, err := wc.UserService.GetUserByID(requesterUserID)
	if err != nil {
		logrus.Errorf("❌ [WorkerController] Failed to get requester from DB: %v", err)
		utils.Error(w, http.StatusUnauthorized, "Invalid user")
		return
	}
	logrus.Infof("🏢 [WorkerController] Requester DB: Name=%s, CarWashID=%v", requester.Name, requester.CarWashID)

	// Parse worker data from request body
	var workerInput models.User
	if err := json.NewDecoder(r.Body).Decode(&workerInput); err != nil {
		logrus.Errorf("❌ [WorkerController] JSON Decode failed: %v", err)
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}
	logrus.Infof("👷 [WorkerController] Decoded Worker Input: Name=%s, Email=%s, CarWashID=%v", workerInput.Name, workerInput.Email, workerInput.CarWashID)

	// Create worker using service
	err = wc.WorkerService.CreateWorker(*requester, workerInput)
	if err != nil {
		logrus.Errorf("❌ Failed to create worker: %v", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.Infof("✅ Worker created: %s", workerInput.Email)
	utils.JSON(w, http.StatusCreated, map[string]string{
		"message": "Worker created successfully",
	})
}

func (wc *WorkerController) GetWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	carwashID := params["id"]

	workers, err := wc.WorkerService.GetWorkersByCarwashID(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, workers)
}

// UpdateWorkerStatus handles status update (e.g., active, on_break, busy)
func (wc *WorkerController) UpdateWorkerStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workerID := params["id"]

	var data struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateWorkerStatus(data.Status); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid status: "+err.Error())
		return
	}

	if err := wc.WorkerService.SetWorkerStatus(workerID, data.Status); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Worker status updated"})
}

// GetAvailableWorkersForBusiness handles getting available workers for assignment
func (wc *WorkerController) GetAvailableWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	businessID := params["id"]

	workers, err := wc.WorkerService.GetAvailableWorkersForAssignment(businessID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, workers)
}

// UpdateWorkerWorkStatus handles worker operational status update (online, offline, busy, on_break)
func (wc *WorkerController) UpdateWorkerWorkStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workerID := params["id"]

	var data struct {
		WorkStatus string `json:"work_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := wc.WorkerService.SetWorkerWorkStatus(workerID, data.WorkStatus); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Worker work status updated"})
}

// AssignWorkerToOrder handles manual worker assignment to orders
func (wc *WorkerController) AssignWorkerToOrder(w http.ResponseWriter, r *http.Request) {
	var data struct {
		WorkerID string `json:"worker_id"`
		OrderID  string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if data.WorkerID == "" || data.OrderID == "" {
		utils.Error(w, http.StatusBadRequest, "worker_id and order_id are required")
		return
	}

	logrus.Infof("📝 [WorkerController.AssignWorkerToOrder] Assigning Worker %s to Order %s", data.WorkerID, data.OrderID)

	if err := wc.WorkerService.AssignWorkerToOrder(data.WorkerID, data.OrderID); err != nil {
		logrus.Errorf("❌ [WorkerController.AssignWorkerToOrder] Failed: %v", err)
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Infof("✅ [WorkerController.AssignWorkerToOrder] Success")

	utils.JSON(w, http.StatusOK, map[string]string{
		"message":   "Worker successfully assigned to order",
		"worker_id": data.WorkerID,
		"order_id":  data.OrderID,
	})
}

// RemoveWorkerFromOrder handles removing/unassigning worker from order
func (wc *WorkerController) RemoveWorkerFromOrder(w http.ResponseWriter, r *http.Request) {
	var data struct {
		WorkerID string `json:"worker_id"`
		OrderID  string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if data.WorkerID == "" || data.OrderID == "" {
		utils.Error(w, http.StatusBadRequest, "worker_id and order_id are required")
		return
	}

	if err := wc.WorkerService.RemoveWorkerFromOrder(data.WorkerID, data.OrderID); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message":   "Worker successfully removed from order",
		"worker_id": data.WorkerID,
		"order_id":  data.OrderID,
	})
}

// UpdateWorkerHandler handles worker details update
func (wc *WorkerController) UpdateWorkerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workerID := vars["id"]

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	err := wc.WorkerService.UpdateWorker(workerID, updateData)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Worker profile updated successfully",
	})
}

// UploadWorkerPhotoHandler handles worker photo upload
func (wc *WorkerController) UploadWorkerPhotoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workerID := vars["id"]

	// 1. Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "File too large or invalid multipart")
		return
	}

	// 2. Get file from form
	file, header, err := r.FormFile("photo")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Missing 'photo' file in request")
		return
	}
	defer file.Close()

	// 3. Create photo file struct
	photoFile := &services.ProfilePhotoFile{
		File:     file,
		Filename: header.Filename,
		Size:     header.Size,
	}

	// 4. Call service
	url, err := wc.WorkerService.UploadWorkerPhoto(workerID, photoFile)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. Return success
	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Worker photo uploaded successfully",
		"url":     url,
	})
}
