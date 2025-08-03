package controllers

import (

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"

	
)

//  1. CreateCarHandler - POST /api/cars/
func CreateCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	accountType := authCtx.AccountType
	role := authCtx.Role 
	fmt.Println("user_id: ", userID)

	// ownerID, err := primitive.ObjectIDFromHex(userID)

	var input models.Car
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := input.Validate(); err != nil {
	utils.Error(w, http.StatusBadRequest, err.Error())
	return
    }
	 
	if role != "car_owner" || accountType != "car_owner" {
		utils.Error(w, http.StatusForbidden, "Only car owners can create cars")
		return
	}
	

	newCar, err := services.CreateCar(userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, newCar)
}

//  2. GetMyCarsHandler - GET /api/cars/my
func GetMyCarsHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	

	cars, err := services.GetCarsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, cars)
}

//  3. UpdateCarHandler - PUT /api/cars/{carID}
func UpdateCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	carID := mux.Vars(r)["carID"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update data")
		return
	}

	err := services.UpdateCar(userID, carID, updates)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Car updated successfully"})
}

//  4. DeleteCarHandler - DELETE /api/cars/{carID}
func DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	carID := mux.Vars(r)["carID"]

	err := services.DeleteCar(userID, carID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Car deleted successfully"})
}

//  5. SetDefaultCarHandler - PATCH /api/cars/{carID}/default
func SetDefaultCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	carID := mux.Vars(r)["carID"]

	err := services.SetDefaultCar(userID, carID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Default car set successfully"})
}



