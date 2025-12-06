package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new UserController instance
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUserProfile handles GET /api/user/{id}
func (uc *UserController) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Validate ObjectID
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logrus.Error("Invalid user ID format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 3. Call service to fetch user
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		logrus.Error("Failed to fetch user: ", err)
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	// 4. Return user profile
	utils.JSON(w, http.StatusOK, user)
}

// UpdateUserProfile handles PUT /api/user/{id}
func (uc *UserController) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	logrus.Info("UpdateUserProfile hit")

	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Validate ObjectID
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logrus.Error("Invalid user ID format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 3. Check Content-Type
	contentType := r.Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

	if isMultipart {
		// Parse multipart form (32MB max memory)
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			logrus.Error("Failed to parse multipart form: ", err)
			utils.Error(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}

		// Extract form fields
		input := &models.User{
			Name:  r.FormValue("name"),
			Phone: r.FormValue("phone"),
			Email: r.FormValue("email"),
		}

		// Validate input
		if err := validation.ValidateStruct(input,
			validation.Field(&input.Name, validation.Length(2, 50)),
			validation.Field(&input.Phone, validation.Length(10, 15)),
			validation.Field(&input.Email),
		); err != nil {
			logrus.Error("Validation failed: ", err)
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Extract file if present
		file, header, err := r.FormFile("profile_photo")
		var profileFile *services.ProfilePhotoFile
		if err == nil {
			defer file.Close()
			profileFile = &services.ProfilePhotoFile{
				File:     file,
				Filename: header.Filename,
				Size:     header.Size,
			}
		} else if err != http.ErrMissingFile {
			logrus.Error("Error processing uploaded file: ", err)
			utils.Error(w, http.StatusBadRequest, "Error processing uploaded file")
			return
		}

		// Call service with file data
		updatedUser, err := uc.userService.UpdateUserWithPhoto(userID, input, profileFile)
		if err != nil {
			logrus.Error("Failed to update user with photo: ", err)
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Return updated user
		utils.JSON(w, http.StatusOK, updatedUser)
	} else if contentType == "application/json" {
		// Handle JSON request
		var input models.User
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			logrus.Error("Failed to parse JSON: ", err)
			utils.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		defer r.Body.Close()

		// Validate input
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Name, validation.Length(2, 50)),
			validation.Field(&input.Phone, validation.Length(10, 15)),
			validation.Field(&input.Email),
			validation.Field(&input.Role, validation.In("car_owner", "business_owner")),
			validation.Field(&input.AccountType, validation.In("car_owner", "car_wash")),
		); err != nil {
			logrus.Error("Validation failed: ", err)
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Call service (no file)
		updatedUser, err := uc.userService.UpdateUser(userID, &input)
		if err != nil {
			logrus.Error("Failed to update user: ", err)
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.JSON(w, http.StatusOK, updatedUser)
	} else {
		logrus.Error("Unsupported Content-Type: ", contentType)
		utils.Error(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json or multipart/form-data")
		return
	}
}

// DeleteUser handles DELETE /api/user/{id}
func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Call service to delete user
	err := uc.userService.DeleteUser(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return success message
	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// GetUserRole handles GET /api/user/{id}/role
func (uc *UserController) GetUserRole(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Get the user's role
	role, err := uc.userService.GetUserRole(userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	// 3. Return the role
	utils.JSON(w, http.StatusOK, map[string]string{
		"role": role,
	})
}

// GetLoyaltyPoints handles GET /api/user/{id}/loyalty
func (uc *UserController) GetLoyaltyPoints(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	points, err := uc.userService.GetLoyaltyPoints(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]int{
		"points": points,
	})
}

// GetPublicUser handles GET /api/user/{id}/public
func (uc *UserController) GetPublicUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	profile, err := uc.userService.GetPublicProfile(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, profile)
}

// GetCurrentUser handles GET /api/user/callback/me
func (uc *UserController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	logrus.Info("👉 /api/user/me endpoint hit")

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		logrus.Warn("❌ Failed to extract user ID from request: ", err)
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	logrus.Infof("✅ Extracted userID from token: %s", userID)

	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		logrus.Warnf("❌ User not found in DB for userID=%s: %v", userID, err)
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	logrus.Infof("✅ Found user in DB: %+v", user)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// AddUserAddress handles POST /api/user/{id}/addresses
func (uc *UserController) AddUserAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	logrus.Infof("Attempting to add address for user ID: %s", userID)

	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 2. Parse address from request body
	var address models.UserAddress
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	// 3. Call service to add address
	if err := uc.userService.AddAddress(objID, address); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{
		"message": "Address added successfully",
	})
}

// DeleteUserAddress handles DELETE /api/user/{id}/addresses/{address_id}
func (uc *UserController) DeleteUserAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	addressID := params["address_id"]
	logrus.Infof("Attempting to delete address ID: %s for user ID: %s", addressID, userID)

	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 2. Call service to delete address
	if err := uc.userService.DeleteAddress(objID, addressID); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Address deleted successfully",
	})
}

// UploadProfilePhoto handles POST /api/user/{id}/photo
func (uc *UserController) UploadProfilePhoto(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Validate ObjectID
	if _, err := primitive.ObjectIDFromHex(userID); err != nil {
		logrus.Error("Invalid user ID format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 3. Parse multipart form (10MB max memory)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logrus.Error("Failed to parse multipart form: ", err)
		utils.Error(w, http.StatusBadRequest, "Could not parse multipart form")
		return
	}

	// 4. Extract file from request
	file, header, err := r.FormFile("profile_photo")
	if err != nil {
		if err == http.ErrMissingFile {
			logrus.Error("No file provided in request for profile_photo")
			utils.Error(w, http.StatusBadRequest, "profile_photo file is required")
		} else {
			logrus.Error("Error processing uploaded file: ", err)
			utils.Error(w, http.StatusBadRequest, "Error processing uploaded file")
		}
		return
	}
	defer file.Close()

	// 5. Construct ProfilePhotoFile for service layer
	photoFile := &services.ProfilePhotoFile{
		File:     file,
		Filename: header.Filename,
		Size:     header.Size,
	}

	// 6. Call service to handle upload and user update
	updatedUser, err := uc.userService.UpdateUserWithPhoto(userID, &models.User{}, photoFile)
	if err != nil {
		logrus.Error("Failed to update user with photo: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 7. Return updated user profile
	utils.JSON(w, http.StatusOK, updatedUser)
}

// DeleteProfilePhoto handles DELETE /api/user/{id}/photo
func (uc *UserController) DeleteProfilePhoto(w http.ResponseWriter, r *http.Request) {
	// 1. Extract user ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Validate ObjectID
	if _, err := primitive.ObjectIDFromHex(userID); err != nil {
		logrus.Error("Invalid user ID format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 3. Call service to handle photo deletion and user update
	updatedUser, err := uc.userService.DeleteProfilePhoto(userID)
	if err != nil {
		logrus.Error("Failed to delete user profile photo: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Return updated user profile
	utils.JSON(w, http.StatusOK, updatedUser)
}

// GetUserAddresses handles GET /api/user/{id}/addresses
func (uc *UserController) GetUserAddresses(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	addresses, err := uc.userService.GetUserAddresses(objID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, addresses)
}

// UpdateUserAddress handles PUT /api/user/{id}/addresses/{address_id}
func (uc *UserController) UpdateUserAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	addressID := params["address_id"]

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := uc.userService.UpdateAddress(objID, addressID, updates); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Address updated successfully",
	})
}

// SetDefaultAddress handles PUT /api/user/{id}/addresses/{address_id}/default
func (uc *UserController) SetDefaultAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	addressID := params["address_id"]

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := uc.userService.SetDefaultAddress(objID, addressID); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Default address set successfully",
	})
}