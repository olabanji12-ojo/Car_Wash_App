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


