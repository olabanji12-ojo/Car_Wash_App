package controllers

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
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

// PUT /api/user/{id} - UPDATED TO HANDLE FILE UPLOADS
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	params := mux.Vars(r)
	userID := params["id"]

	// 2. Check Content-Type to determine how to parse request
	contentType := r.Header.Get("Content-Type")
	
	// If it's multipart form data (file upload), handle differently
	if contentType != "" && contentType[:19] == "multipart/form-data" {
		// Parse multipart form (32MB max memory)
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}

		// Extract form fields
		input := &models.User{
			Name:  r.FormValue("name"),
			Phone: r.FormValue("phone"),
		}
		
		// Note: Email might be read-only in updates, depends on your business logic
		if email := r.FormValue("email"); email != "" {
			input.Email = email
		}

		// Extract file if present
		file, header, err := r.FormFile("profile_photo")
		var profileFile *services.ProfilePhotoFile
		
		if err == nil {
			// File was uploaded successfully
			defer file.Close()
			profileFile = &services.ProfilePhotoFile{
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

		// 3. Call service with file data
		updatedUser, err := services.UpdateUserWithPhoto(userID, input, profileFile)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		// 4. Return updated user
		utils.JSON(w, http.StatusOK, updatedUser)

	} else {
		// Handle as JSON (backward compatibility)
		var input *models.User
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Call original service (no file)
		updatedUser, err := services.UpdateUser(userID, input)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.JSON(w, http.StatusOK, updatedUser)
	}
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