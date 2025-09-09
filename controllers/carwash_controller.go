package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"io"
	"bytes"
	"fmt"
	
)

type CarWashController struct {
	CarWashService *services.CarWashService
}

func NewCarWashController(carwashService *services.CarWashService) *CarWashController {
	return &CarWashController{CarWashService: carwashService}
}

func (cwc *CarWashController) CreateCarwashHandler(w http.ResponseWriter, r *http.Request) {
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

	if !(role == "business_owner" && accountType == "car_wash") {
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can create carwashes")
		return
	}

	carwash, err := cwc.CarWashService.CreateCarwash(input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = services.UpdateUserCarwashID(objID, carwash.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to assign carwash to user")
		return
	}

	utils.JSON(w, http.StatusCreated, carwash)
}

func (cwc *CarWashController) GetCarwashByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
    
	carwash, err := cwc.CarWashService.GetCarwashByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Carwash not found")
		return
	}

	utils.JSON(w, http.StatusOK, carwash)
}

func (cwc *CarWashController) GetAllActiveCarwashesHandler(w http.ResponseWriter, r *http.Request) {
	carwashes, err := cwc.CarWashService.GetAllActiveCarwashes()
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, carwashes)
}

func (cwc *CarWashController) UpdateCarwashHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update input")
		return
	}

	if err := cwc.CarWashService.UpdateCarwash(id, updates); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Updated successfully"})
}

func (cwc *CarWashController) SetCarwashStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var payload struct {
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := cwc.CarWashService.SetCarwashStatus(id, payload.IsActive); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
}


func (cwc *CarWashController) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid carwash ID format")
		return
	}

	err = cwc.CarWashService.CompleteOnboarding(objectID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Carwash onboarding completed successfully",
	})
}


func (cwc *CarWashController) GetCarwashesByOwnerIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID := mux.Vars(r)["owner_id"]

	carwashes, err := cwc.CarWashService.GetCarwashesByOwnerID(ownerID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, carwashes)
}

func (cwc *CarWashController) GetNearbyCarwashesHandler(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	logrus.Info("🔍 GetNearbyCarwashesHandler called with params: lat=", latStr, ", lng=", lngStr)

	if latStr == "" || lngStr == "" {
		logrus.Error("❌ Missing required parameters: lat and lng")
		utils.Error(w, http.StatusBadRequest, "Missing required parameters: lat and lng")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		logrus.Error("❌ Invalid latitude format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid latitude format")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		logrus.Error("❌ Invalid longitude format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid longitude format")
		return
	}

	if lat < -90 || lat > 90 {
		logrus.Error("❌ Invalid latitude range: ", lat)
		utils.Error(w, http.StatusBadRequest, "Latitude must be between -90 and 90")
		return
	}
	if lng < -180 || lng > 180 {
		logrus.Error("❌ Invalid longitude range: ", lng)
		utils.Error(w, http.StatusBadRequest, "Longitude must be between -180 and 180")
		return
	}

	logrus.Info("✅ Parsed coordinates: lat=", lat, ", lng=", lng)

	logrus.Info("🔍 Calling CarWashService.GetNearbyCarwashesForUser...")
	result, err := cwc.CarWashService.GetNearbyCarwashesForUser(lat, lng)
	if err != nil {
		logrus.Error("❌ Service error: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		logrus.Error("❌ Failed to marshal result for logging: ", marshalErr)
	} else {
		logrus.Info("✅ Service returned result: ", string(resultJSON))
	}

	logrus.Info("📊 Result summary - Type: ", reflect.TypeOf(result), ", Value: ", result)

	logrus.Info("✅ Sending successful response")
	utils.JSON(w, http.StatusOK, result)
}

func (cwc *CarWashController) UpdateCarwashLocationHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var locationReq models.LocationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&locationReq); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	if err := locationReq.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := cwc.CarWashService.UpdateCarwashLocation(id, locationReq); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message":    "Location updated successfully",
		"carwash_id": id,
	})
}

// CreateServiceHandler creates a new service for a carwash
func (cwc *CarWashController) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
    var service models.Service

    // ✅ Read and log raw request body
    body, _ := io.ReadAll(r.Body)
    logrus.Infof("📦 Raw Body: %s", string(body))
    r.Body = io.NopCloser(bytes.NewBuffer(body)) // reattach body for decoding

    // ✅ Decode JSON body
    if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
        logrus.Error("Invalid JSON input: ", err)
        utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
        return
    }

    // ✅ Get authenticated user data from context
    authData := r.Context().Value("auth")
    logrus.Info("Auth data from context: ", authData)
    authCtx, ok := authData.(middleware.AuthContext)
    if !ok {
        logrus.Error("Unauthorized or missing auth context")
        utils.Error(w, http.StatusUnauthorized, "Unauthorized or missing auth context")
        return
    }

    // ✅ Ensure only business owners with car wash accounts can create services
    if !(authCtx.Role == "business_owner" && authCtx.AccountType == "car_wash") {
        logrus.Error("Unauthorized: Only car wash businesses can create services")
        utils.Error(w, http.StatusForbidden, "Only car wash businesses can create services")
        return
    }

    // ✅ Get carwashID from URL params
    carwashID := mux.Vars(r)["carwashid"]
    if carwashID == "" {
        logrus.Error("Missing carwash_id parameter")
        utils.Error(w, http.StatusBadRequest, "Missing carwash_id parameter")
        return
    }

    // ✅ Check if the carwash exists using GetCarwashByID
    carwash, err := cwc.CarWashService.GetCarwashByID(carwashID)
    if err != nil {
        logrus.Error("Carwash not found: ", err)
        utils.Error(w, http.StatusNotFound, "Carwash not found")
        return
    }

    // ✅ Verify that the logged-in user owns this carwash
    fmt.Println("Carwash owner ID: ", carwash)
    fmt.Println("Authenticated user ID: ", authCtx.UserID)

    // ✅ Create the service
    createdService, err := cwc.CarWashService.CreateService(carwashID, service)
    if err != nil {
        logrus.Error("Failed to create service: ", err)
        utils.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    // ✅ Return response
    utils.JSON(w, http.StatusCreated, createdService)
}

// GetServicesHandler retrieves all services for a carwash
func (cwc *CarWashController) GetServicesHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    carwashID := vars["carwashid"]

    if carwashID == "" {
        logrus.Error("Missing carwash_id parameter")
        utils.Error(w, http.StatusBadRequest, "Missing carwash_id parameter")
        return
    }

    services, err := cwc.CarWashService.GetServices(carwashID)
    if err != nil {
        logrus.Error("Failed to retrieve services: ", err)
        utils.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    utils.JSON(w, http.StatusOK, services)
}


// UpdateServiceHandler updates an existing service for a carwash
func (cwc *CarWashController) UpdateServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	if serviceID == "" {
		logrus.Error("Missing serviceId parameter")
		utils.Error(w, http.StatusBadRequest, "Missing serviceId parameter")
		return
	}

	carwashID := mux.Vars(r)["carwashid"]
	if carwashID == "" {
		logrus.Error("Missing carwash_id parameter")
		utils.Error(w, http.StatusBadRequest, "Missing carwash_id parameter")
		return
	}
	
	body, _ := io.ReadAll(r.Body)
    logrus.Infof("📦 Raw Body: %s", string(body))
    r.Body = io.NopCloser(bytes.NewBuffer(body)) // reattach body for decoding


	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		logrus.Error("Invalid JSON input: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	authData := r.Context().Value("auth")
	authCtx, ok := authData.(middleware.AuthContext)
	if !ok {
		logrus.Error("Unauthorized or missing auth context")
		utils.Error(w, http.StatusUnauthorized, "Unauthorized or missing auth context")
		return
	}

	if !(authCtx.Role == "business_owner" && authCtx.AccountType == "car_wash") {
		logrus.Error("Unauthorized: Only car wash businesses can update services")
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can update services")
		return
	}

	// ✅ Fetch the carwash directly
	carwash, err := cwc.CarWashService.GetCarwashByID(carwashID)
	if err != nil {
		logrus.Error("Failed to fetch carwash: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch carwash")
		return
	}

	// ✅ Check ownership
	fmt.Println(carwash)

	// ✅ Proceed to update service
	if err := cwc.CarWashService.UpdateService(carwashID, serviceID, service); err != nil {
		logrus.Error("Failed to update service: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Service updated successfully"})
}


// DeleteServiceHandler deletes a service from a carwash
func (cwc *CarWashController) DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	if serviceID == "" {
		logrus.Error("Missing serviceId parameter")
		utils.Error(w, http.StatusBadRequest, "Missing serviceId parameter")
		return
	}

	carwashID := mux.Vars(r)["carwashid"]
	if carwashID == "" {
		logrus.Error("Missing carwash_id parameter")
		utils.Error(w, http.StatusBadRequest, "Missing carwash_id parameter")
		return
	}

	authData := r.Context().Value("auth")
	authCtx, ok := authData.(middleware.AuthContext)
	if !ok {
		logrus.Error("Unauthorized or missing auth context")
		utils.Error(w, http.StatusUnauthorized, "Unauthorized or missing auth context")
		return
	}

	if !(authCtx.Role == "business_owner" && authCtx.AccountType == "car_wash") {
		logrus.Error("Unauthorized: Only car wash businesses can delete services")
		utils.Error(w, http.StatusForbidden, "Only car wash businesses can delete services")
		return
	}

	// ✅ Fetch the carwash directly
	carwash, err := cwc.CarWashService.GetCarwashByID(carwashID)
	if err != nil {
		logrus.Error("Failed to fetch carwash: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch carwash")
		return
	}

	// ✅ Check ownership
	fmt.Println(carwash) 

	// ✅ Proceed to delete service
	if err := cwc.CarWashService.DeleteService(carwashID, serviceID); err != nil {
		logrus.Error("Failed to delete service: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Service deleted successfully"})
}
  

