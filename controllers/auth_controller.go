package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"

	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	// "github.com/go-ozzo/ozzo-validation/v4/is"
)

type AuthController struct {
	AuthService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{AuthService: authService}
}

// REGISTER HANDLER

func (ac *AuthController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Info("RegisterHandler hit")

	// Parse multipart form (32MB max)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		logrus.Error("Failed to parse multipart form: ", err)
		utils.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Extract form fields
	input := models.User{
		Name:        r.FormValue("name"),
		Email:       r.FormValue("email"),
		Phone:       r.FormValue("phone"),
		Password:    r.FormValue("password"),
		AccountType: r.FormValue("account_type"),
		Role:        r.FormValue("role"),
	}

	// Validate input
	if err := validateRegistrationInput(input); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Handle profile photo upload
	var profilePhotoURL string
	file, header, err := r.FormFile("profile_photo")
	if err == nil { // File was provided
		defer file.Close()

		// Validate file type
		if !isValidImageType(header.Filename) {
			utils.Error(w, http.StatusBadRequest, "Invalid file type. Only JPG, PNG, GIF allowed")
			return
		}

		// Generate unique filename
		filename := fmt.Sprintf("user_%s_%d",
			strings.ReplaceAll(input.Email, "@", "_"),
			time.Now().Unix())

		// Upload to Cloudinary
		uploadResult, err := services.UploadImage(file, filename, "profile_photos")
		if err != nil {
			logrus.Error("Image upload failed: ", err)
			utils.Error(w, http.StatusInternalServerError, "Failed to upload profile photo")
			return
		}

		profilePhotoURL = uploadResult.SecureURL
		logrus.Info("Image uploaded successfully: ", profilePhotoURL)
	}

	// Set profile photo URL in user struct
	input.ProfilePhoto = profilePhotoURL

	// Call service to register user
	newUser, err := ac.AuthService.RegisterUser(input)
	if err != nil {
		// If user creation fails but image was uploaded, clean up
		if profilePhotoURL != "" {
			go func() {
				// Extract public_id from URL or store it separately
				// For now, we'll skip cleanup, but in production you should handle this
				logrus.Warn("User creation failed but image was uploaded: ", profilePhotoURL)
			}()
		}
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, newUser)
}

// Helper function to validate image file types
func isValidImageType(filename string) bool {
	validTypes := []string{".jpg", ".jpeg", ".png", ".gif"}
	filename = strings.ToLower(filename)

	for _, ext := range validTypes {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// Validation function for registration input
func validateRegistrationInput(input models.User) error {
	return validation.ValidateStruct(&input,
		validation.Field(&input.Name, validation.Required, validation.Length(2, 100)),
		// validation.Field(&input.Email, validation.Required, is.Email),
		validation.Field(&input.Phone, validation.Required),
		validation.Field(&input.Password, validation.Required, validation.Length(6, 100)),
		validation.Field(&input.AccountType, validation.Required, validation.In("car_owner", "car_wash")),
		validation.Field(&input.Role, validation.Required, validation.In("car_owner", "business_owner", "worker", "business_admin")),
	)
}

// LOGIN HANDLER

func (ac *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode JSON login input
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logrus.Warn("Invalid login input: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	//  Validate login input
	err := validation.ValidateStruct(&credentials,
		// validation.Field(&credentials.Email, validation.Required, is.Email),
		validation.Field(&credentials.Password, validation.Required, validation.Length(6, 100)),
	)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call service to login
	token, user, err := ac.AuthService.LoginUser(credentials.Email, credentials.Password)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	utils.JSON(w, http.StatusOK, response)
}
