package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

	"fmt"

)

//  POST /api/carwashes — Create new carwash
func CreateCarwashHandler(w http.ResponseWriter, r *http.Request) {
	var input models.Carwash

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}


	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	ownerID  := authCtx.UserID
	role := authCtx.Role
	fmt.Println("Owner_id: ", ownerID)

	if role == "car_owner" {
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can create carwashes")
		return
	}

	if ownerID == "" {
		utils.Error(w, http.StatusUnauthorized, "Missing owner ID")
		return
	}

// 	if err := input.Validate(); err != nil {
// 	utils.Error(w, http.StatusBadRequest, err.Error())
// 	return
//    }

	carwash, err := services.CreateCarwash(ownerID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, carwash)
}

//  GET /api/carwashes/{id} — View carwash profile by ID
func GetCarwashByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	carwash, err := services.GetCarwashByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Carwash not found")
		return
	}

	utils.JSON(w, http.StatusOK, carwash)
}

//  GET /api/carwashes — View all active carwashes
func GetAllActiveCarwashesHandler(w http.ResponseWriter, r *http.Request) {
	carwashes, err := services.GetAllActiveCarwashes()
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, carwashes)
}

//  PUT /api/carwashes/{id} — Update carwash profile
func UpdateCarwashHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update input")
		return
	}

	if err := services.UpdateCarwash(id, updates); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Updated successfully"})
}

//  PATCH /api/carwashes/{id}/status — Toggle is_active status
func SetCarwashStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var payload struct {
		
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := services.SetCarwashStatus(id, payload.IsActive); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
}

//  GET /api/carwashes/owner/{owner_id} — Business can view their own carwashes
func GetCarwashesByOwnerIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID := mux.Vars(r)["owner_id"]

	carwashes, err := services.GetCarwashesByOwnerID(ownerID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, carwashes)
}


