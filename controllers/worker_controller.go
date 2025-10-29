package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
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
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	requesterUserID := authCtx.UserID

	// Get requester (business owner) details
	requester, err := wc.UserService.GetUserByID(requesterUserID)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Parse worker data from request body
	var workerInput models.User
	if err := json.NewDecoder(r.Body).Decode(&workerInput); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	// Create worker using service
	err = wc.WorkerService.CreateWorker(*requester, workerInput)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{
		"message": "Worker created successfully",
	})
}

func (wc *WorkerController) GetWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	businessID := params["id"]

	workers, err := wc.WorkerService.GetWorkersByBusinessID(businessID)
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

	if err := wc.WorkerService.AssignWorkerToOrder(data.WorkerID, data.OrderID); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

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