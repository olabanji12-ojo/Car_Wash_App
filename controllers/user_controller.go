package controllers

import (
	"net/http"

	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	// "github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/models"
)



// GET /api/user/{id}
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Call service to fetch user
	user, err := services.GetUserByID(userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	// 3. Return user profile (password already stripped in service)
	utils.JSON(w, http.StatusOK, user)
}


// PUT /api/user/{id}
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Decode the JSON body
	var input *models.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 3. Call service to update user
	updatedUser, err := services.UpdateUser(userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Return updated user
	utils.JSON(w, http.StatusOK, updatedUser)
}


// DELETE /api/user/{id}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Call service to delete user
	err := services.DeleteUser(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return success message
	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}


// GET /api/user/{id}/role
func GetUserRole(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Get the user's role
	role, err := services.GetUserRole(userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	// 3. Return the role
	utils.JSON(w, http.StatusOK, map[string]string{
		"role": role,
	})
}


func GetLoyaltyPoints(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	points, err := services.GetLoyaltyPoints(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]int{
		"points": points,
	})
}


// func GetWorkersForBusiness(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	businessID := params["id"]

// 	workers, err := services.GetWorkersByBusinessID(businessID)
// 	if err != nil {
// 		utils.Error(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	utils.JSON(w, http.StatusOK, workers)
// }


func GetPublicUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	profile, err := services.GetPublicProfile(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, profile)
}














