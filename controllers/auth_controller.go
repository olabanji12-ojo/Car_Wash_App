package controllers

import (
	"encoding/json"
	"net/http"

	"crypto/rand"
    "encoding/base64"
	"context"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/sirupsen/logrus"
    "net/url"
	"os"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
    

	"fmt"
	"strings"
	"time"
	"io/ioutil"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	// "github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/olabanji12-ojo/CarWashApp/repositories"
    "github.com/olabanji12-ojo/CarWashApp/database"
    "go.mongodb.org/mongo-driver/bson"
	// "github.com/joho/godotenv"
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


// GOOGLE LOGIN HANDLER
// helper to create secure random state and set cookie
// generate a random nonce and store it in a cookie
func generateNonceCookie(w http.ResponseWriter) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	tokenString := base64.URLEncoding.EncodeToString(b)

	cookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true, // set true in production (HTTPS)
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	return tokenString, nil
}

// GoogleLoginHandler expects optional query params: ?role=car_owner&account_type=car_owner
func (ac *AuthController) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1) create nonce cookie
	nonce, err := generateNonceCookie(w)
	if err != nil {
		logrus.Error("failed to generate nonce: ", err)
		utils.Error(w, http.StatusInternalServerError, "server error")
		return
	}

	// 2) read role/account_type from query params (frontend should pass them)
	role := r.URL.Query().Get("role")
	accountType := r.URL.Query().Get("account_type")

	// 3) build payload and encode it
	values := url.Values{}
	if role != "" {
		values.Set("role", role)
	}
	if accountType != "" {
		values.Set("account_type", accountType)
	}
	payload := values.Encode() // e.g. "role=car_owner&account_type=car_owner"
	encodedPayload := base64.URLEncoding.EncodeToString([]byte(payload))

	// 4) build state = nonce|encodedPayload
	state := nonce + "|" + encodedPayload

	// 5) redirect to Google with the state
	authURL := config.GoogleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}



func (ac *AuthController) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// 1) parse state and validate nonce cookie
	queryState := r.URL.Query().Get("state")
	if queryState == "" {
		utils.Error(w, http.StatusBadRequest, "missing state")
		return
	}
	parts := strings.SplitN(queryState, "|", 2)
	if len(parts) != 2 {
		utils.Error(w, http.StatusBadRequest, "invalid state format")
		return
	}
	nonceFromState := parts[0]
	encodedPayload := parts[1]

	cookie, err := r.Cookie("oauthstate")
	if err != nil || cookie.Value == "" || cookie.Value != nonceFromState {
		logrus.Warn("invalid or missing oauthstate cookie")
		utils.Error(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	// 2) decode payload to get role/account_type
	payloadBytes, err := base64.URLEncoding.DecodeString(encodedPayload)
	role := ""
	accountType := ""
	if err == nil && len(payloadBytes) > 0 {
		values, _ := url.ParseQuery(string(payloadBytes))
		role = values.Get("role")
		accountType = values.Get("account_type")
	}
	// If frontend didn't pass values, you can set a fallback or leave empty.
	if role == "" {
		// fallback (optional). You can change this to reject or force frontend to pass.
		role = "car_owner"
	}
	if accountType == "" {
		accountType = "car_owner"
	}

	// 3) get code and exchange for token
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.Error(w, http.StatusBadRequest, "code not found")
		return
	}

	ctx := context.Background()
	token, err := config.GoogleOauthConfig.Exchange(ctx, code)
	if err != nil {
		logrus.Error("token exchange failed: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	// 4) fetch userinfo from Google
	client := config.GoogleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logrus.Error("failed to fetch userinfo: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("failed to read userinfo body: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to read user info")
		return
	}

	var gi struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &gi); err != nil {
		logrus.Error("failed to parse userinfo: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to parse user info")
		return
	}

	// 5) Upsert user (use your repo)
	userRepo := repositories.NewUserRepository(database.DB)
	existingUser, err := userRepo.FindUserByEmail(gi.Email)
	var u *models.User
	now := time.Now()
	if err == nil && existingUser != nil {
		// update fields:
		update := bson.M{
			"name":          gi.Name,
			"profile_photo": gi.Picture,
			"verified":      true,
			"updated_at":    now,
		}
		if err := userRepo.UpdateUserByID(existingUser.ID, update); err != nil {
			logrus.Error("failed to update existing user: ", err)
			utils.Error(w, http.StatusInternalServerError, "db error")
			return
		}
		u, _ = userRepo.FindUserByEmail(gi.Email)
	} else {
		// create new user using role/account_type from state payload
		newUser := models.User{
			Name:         gi.Name,
			Email:        gi.Email,
			ProfilePhoto: gi.Picture,
			Verified:     true,
			Role:         role,
			AccountType:  accountType,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := userRepo.CreateUser(newUser); err != nil {
			logrus.Error("failed to create user: ", err)
			utils.Error(w, http.StatusInternalServerError, "db error")
			return
		}
		u, _ = userRepo.FindUserByEmail(gi.Email)
	}

	// 6) create your JWT and redirect to frontend
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logrus.Error("missing JWT_SECRET env")
		utils.Error(w, http.StatusInternalServerError, "server misconfigured")
		return
	}
	claims := jwt.MapClaims{
		"user_id": u.ID.Hex(),
		"email":   u.Email,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := tokenJwt.SignedString([]byte(jwtSecret))
	if err != nil {
		logrus.Error("failed to sign jwt: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    signedToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // true in production (HTTPS)
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(72 * time.Hour),
	})
	  
	// redirect to frontend without token in URL
	callbackURL := fmt.Sprintf("%s/CallbackPage", os.Getenv("FRONTEND_URL"))
    http.Redirect(w, r, callbackURL, http.StatusFound)

}
