package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/sirupsen/logrus"
	// "fmt"

)

//  POST /api/carwashes — Create new carwash
func CreateCarwashHandler(w http.ResponseWriter, r *http.Request) {
	var input models.Carwash

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}
    
    
	authData := r.Context().Value("auth")
	logrus.Info("Auth data from context:", authData) 
	authCtx, ok := authData.(middleware.AuthContext)

	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized or missing auth context")
		return
	}

	
	role := authCtx.Role
	accountType := authCtx.AccountType
	userID := authCtx.UserID
	objID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	
 
	if !(role == "business_owner" && accountType == "car_wash"){
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can create carwashes")
		return
	}
 	
	
// 	if err := input.Validate(); err != nil {
// 	utils.Error(w, http.StatusBadRequest, err.Error())
// 	return
//    }

	carwash, err := services.CreateCarwash(input)
	 
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update the user’s carwash_id field
	err = services.UpdateUserCarwashID(objID, carwash.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to assign carwash to user")
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


// GET /api/carwashes/nearby?lat={lat}&lng={lng} — Find nearby carwashes with fallback
func GetNearbyCarwashesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse latitude and longitude from query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	if latStr == "" || lngStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing required parameters: lat and lng")
		return
	}
    
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid latitude format")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid longitude format")
		return
	}

	// Validate latitude and longitude ranges
	if lat < -90 || lat > 90 {
		utils.Error(w, http.StatusBadRequest, "Latitude must be between -90 and 90")
		return
	}
	if lng < -180 || lng > 180 {
		utils.Error(w, http.StatusBadRequest, "Longitude must be between -180 and 180")
		return
	}

	// Call service to get nearby carwashes
	result, err := services.GetNearbyCarwashesForUser(lat, lng)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// PUT /api/carwashes/{id}/location — Update carwash location and service range
func UpdateCarwashLocationHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var locationReq models.LocationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&locationReq); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	// Validate the location request
	if err := locationReq.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update carwash location
	if err := services.UpdateCarwashLocation(id, locationReq); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Location updated successfully",
		"carwash_id": id,
	})
}

