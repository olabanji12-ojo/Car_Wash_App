package controllers

import (
	"net/http"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	// "github.com/olabanji12-ojo/CarWashApp/models"
	// "github.com/olabanji12-ojo/CarWashApp/models"
)

func GetWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	businessID := params["id"]

	workers, err := services.GetWorkersByBusinessID(businessID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, workers)

}

// UpdateWorkerStatus handles status update (e.g., active, on_break, busy)
func UpdateWorkerStatus(w http.ResponseWriter, r *http.Request) {
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

	if err := services.SetWorkerStatus(workerID, data.Status); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Worker status updated"})
}

// GetAvailableWorkersForBusiness handles getting available workers for assignment
func GetAvailableWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	businessID := params["id"]

	workers, err := services.GetAvailableWorkersForAssignment(businessID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, workers)
}

// UpdateWorkerWorkStatus handles worker operational status update (online, offline, busy, on_break)
func UpdateWorkerWorkStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workerID := params["id"]

	var data struct {
		WorkStatus string `json:"work_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := services.SetWorkerWorkStatus(workerID, data.WorkStatus); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Worker work status updated"})
}

// AssignWorkerToOrder handles manual worker assignment to orders
func AssignWorkerToOrder(w http.ResponseWriter, r *http.Request) {
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

	if err := services.AssignWorkerToOrder(data.WorkerID, data.OrderID); err != nil {
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
func RemoveWorkerFromOrder(w http.ResponseWriter, r *http.Request) {
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

	if err := services.RemoveWorkerFromOrder(data.WorkerID, data.OrderID); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message":   "Worker successfully removed from order",
		"worker_id": data.WorkerID,
		"order_id":  data.OrderID,
	})
}
