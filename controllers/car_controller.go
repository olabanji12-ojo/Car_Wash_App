package controllers

import (
	"encoding/json"
	// "fmt"
	"strings"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
	
)

type CarController struct {
	CarService *services.CarService
}

func NewCarController(carService *services.CarService) *CarController {
	return &CarController{CarService: carService}
}

// 1. CreateCarHandler - POST /api/cars/ - UPDATED TO HANDLE IMAGE UPLOAD
func (cs *CarController)CreateCarHandler(w http.ResponseWriter, r *http.Request) {
    // Get auth context safely
    authCtxVal := r.Context().Value("auth")
    authCtx, ok := authCtxVal.(middleware.AuthContext)
    if !ok {
        utils.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    userID := authCtx.UserID
    accountType := authCtx.AccountType
    role := authCtx.Role

    if role != "car_owner" || accountType != "car_owner" {
        utils.Error(w, http.StatusForbidden, "Only car owners can create cars")
        return
    }

    contentType := r.Header.Get("Content-Type")

    if strings.HasPrefix(contentType, "multipart/form-data") {
        // Handle multipart/form-data (file upload)
        err := r.ParseMultipartForm(32 << 20) // 32MB
        if err != nil {
            utils.Error(w, http.StatusBadRequest, "Failed to parse form data")
            return
        }

        input := models.Car{
            Model:     r.FormValue("model"),
            Plate:     r.FormValue("plate"),
            Color:     r.FormValue("color"),
            IsDefault: r.FormValue("is_default") == "true",
        }

        if err := input.Validate(); err != nil {
            utils.Error(w, http.StatusBadRequest, err.Error())
            return
        }

        // File upload
        file, header, err := r.FormFile("profile_photo")
        var carPhotoFile *services.CarPhotoFile
        if err == nil {
            defer file.Close()
            carPhotoFile = &services.CarPhotoFile{
                File:     file,
                Filename: header.Filename,
                Size:     header.Size,
            }
        } else if err != http.ErrMissingFile {
            utils.Error(w, http.StatusBadRequest, "Error processing uploaded file")
            return
        }

        newCar, err := cs.CarService.CreateCarWithPhoto(userID, input, carPhotoFile)
        if err != nil {
            utils.Error(w, http.StatusInternalServerError, err.Error())
            return
        }

        utils.JSON(w, http.StatusCreated, newCar)
        return
    }

    // Handle JSON request
    var input models.Car
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        utils.Error(w, http.StatusBadRequest, "Invalid JSON request body")
        return
    }

    if err := input.Validate(); err != nil {
        utils.Error(w, http.StatusBadRequest, err.Error())
        return
    }

    newCar, err := cs.CarService.CreateCar(userID, input)
    if err != nil {
        utils.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    utils.JSON(w, http.StatusCreated, newCar)
}
 

// 2. GetMyCarsHandler - GET /api/cars/my
func (cs *CarController)GetMyCarsHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	cars, err := cs.CarService.GetCarsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, cars)
}

// 3. UpdateCarHandler - PUT /api/cars/{carID} - UPDATED TO HANDLE IMAGE UPLOAD
func (cs *CarController)UpdateCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID
	carID := mux.Vars(r)["carID"]

	// Check Content-Type to determine how to parse request
	contentType := r.Header.Get("Content-Type")
	
	// If it's multipart form data (file upload), handle differently
	if contentType != "" && contentType[:19] == "multipart/form-data" {
		// Parse multipart form (32MB max memory)
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}

		// Extract form fields as updates map
		updates := make(map[string]interface{})
		
		if model := r.FormValue("model"); model != "" {
			updates["model"] = model
		}
		if plate := r.FormValue("plate"); plate != "" {
			updates["plate"] = plate
		}
		if color := r.FormValue("color"); color != "" {
			updates["color"] = color
		}
		if isDefault := r.FormValue("is_default"); isDefault != "" {
			updates["is_default"] = isDefault == "true"
		}

		// Extract file if present
		file, header, err := r.FormFile("profile_photo")
		var carPhotoFile *services.CarPhotoFile
		
		if err == nil {
			// File was uploaded successfully
			defer file.Close()
			carPhotoFile = &services.CarPhotoFile{
				File:     file,
				Filename: header.Filename,
				Size:     header.Size,
			}
		} else if err != http.ErrMissingFile {
			// Error other than missing file
			utils.Error(w, http.StatusBadRequest, "Error processing uploaded file")
			return
		}
		// If err == http.ErrMissingFile, that's fine - no file uploaded

		// Update car with photo
		err = cs.CarService.UpdateCarWithPhoto(userID, carID, updates, carPhotoFile)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.JSON(w, http.StatusOK, map[string]string{"message": "Car updated successfully"})

	} else {
		// Handle as JSON (backward compatibility)
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid update data")
			return
		}

		// Update car without photo (original method)
		err := cs.CarService.UpdateCar(userID, carID, updates)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.JSON(w, http.StatusOK, map[string]string{"message": "Car updated successfully"})
	}
}

// 4. DeleteCarHandler - DELETE /api/cars/{carID}
func (cs *CarController)DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	carID := mux.Vars(r)["carID"]

	err := cs.CarService.DeleteCar(userID, carID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Car deleted successfully"})
}

// 5. SetDefaultCarHandler - PATCH /api/cars/{carID}/default
func (cs *CarController)SetDefaultCarHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	carID := mux.Vars(r)["carID"]
    
	err := cs.CarService.SetDefaultCar(userID, carID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Default car set successfully"})
}

// 6. GetCarByIDHandler - GET /api/cars/{carID}
func (cs *CarController) GetCarByIDHandler(w http.ResponseWriter, r *http.Request) {
	carID := mux.Vars(r)["carID"]
	logrus.Info("Received request for car ID:", carID)

	car, err := cs.CarService.GetCarByID(carID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, car)
}