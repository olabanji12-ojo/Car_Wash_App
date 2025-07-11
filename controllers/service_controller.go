package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

)

//  Create a new service
func CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	role := authCtx.Role

	if role != "business" {
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can create services")
		return
	}

	var input models.Service
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid service data")
		return
	}

	createdService, err := services.CreateService(userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, createdService)
}

//  Get all services for current business user
func GetMyServicesHandler(w http.ResponseWriter, r *http.Request) {

	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	role := authCtx.Role

	if role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can access their services")
		return
	}

	servicesList, err := services.GetServicesByCarwashID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, servicesList)
}

//  Get one service by ID (public access)
func GetServiceByIDHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["id"]

	service, err := services.GetServiceByID(serviceID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, service)
}

//  Update service (business only)
func UpdateServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["id"]
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	role := authCtx.Role

	if role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can update services")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update data")
		return
	}

	if err := services.UpdateService(userID, serviceID, updates); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Service updated successfully"})

}

//  Soft delete service
func DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["id"]
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	role := authCtx.Role

	if role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can delete services")
		return
	}

	if err := services.DeleteService(userID, serviceID); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Service deleted"})
}



