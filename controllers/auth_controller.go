package controllers

import (
	"encoding/json"
	"net/http"

	"context"
	"crypto/rand"
	"encoding/base64"

	"net/url"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	// "github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
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
	logrus.Info("🔵 [RegisterHandler] New registration request received")
	logrus.Infof("🔵 [RegisterHandler] Headers: %+v", r.Header)

	// Log request body
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	logrus.Infof("🔵 [RegisterHandler] Request body: %s", string(bodyBytes))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if r.Header.Get("Content-Type") != "application/json" {
		errMsg := "Content-Type must be application/json"
		logrus.Warnf("❌ [RegisterHandler] %s", errMsg)
		utils.Error(w, http.StatusUnsupportedMediaType, errMsg)
		return
	}

	var input models.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logrus.Errorf("❌ [RegisterHandler] Failed to parse JSON: %v", err)
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	logrus.Infof("🔵 [RegisterHandler] Parsed input: %+v", input)

	if err := validateRegistrationInput(input); err != nil {
		logrus.Warnf("❌ [RegisterHandler] Validation failed: %v", err)
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Info("🔄 [RegisterHandler] Calling AuthService.RegisterUser")
	newUser, err := ac.AuthService.RegisterUser(input)
	if err != nil {
		logrus.Errorf("❌ [RegisterHandler] Registration failed: %v", err)
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Infof("✅ [RegisterHandler] User registered successfully: %s", newUser.Email)

	// Generate JWT token for the new user
	token, err := utils.GenerateToken(newUser.ID.Hex(), newUser.Email, newUser.Role, newUser.AccountType)
	if err != nil {
		logrus.Error("❌ [RegisterHandler] Error generating token: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Create response
	response := map[string]interface{}{
		"message": "Registration successful",
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":           newUser.ID.Hex(),
				"email":        newUser.Email,
				"role":         newUser.Role,
				"account_type": newUser.AccountType,
				"name":         newUser.Name,
				"phone":        newUser.Phone,
			},
			"token": token,
		},
	}

	// Set response headers and encode response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Errorf("❌ [RegisterHandler] Failed to encode response: %v", err)
	}

	logrus.Info("✅ [RegisterHandler] Registration completed and response sent")
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

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logrus.Warn("Invalid login input: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	err := validation.ValidateStruct(&credentials,
		validation.Field(&credentials.Password, validation.Required, validation.Length(6, 100)),
	)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := ac.AuthService.LoginUser(credentials.Email, credentials.Password)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Check if user is verified
	if !user.Verified {
		utils.Error(w, http.StatusForbidden, "Email not verified. Please check your email for verification code.")
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteStrictMode,
	})

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"data": map[string]interface{}{
			"user":  user,
			"token": token,
		},
	})
}

// VERIFY EMAIL HANDLER
func (ac *AuthController) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := ac.AuthService.VerifyEmail(input.Email, input.Token); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// RESEND VERIFICATION EMAIL HANDLER
func (ac *AuthController) ResendVerificationEmailHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := ac.AuthService.ResendVerificationEmail(input.Email); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Verification code resent successfully"})
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
	if err != nil || cookie.Value != nonceFromState {
		logrus.Warn("invalid or missing oauthstate cookie")
		utils.Error(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(encodedPayload)
	role := ""
	accountType := ""
	if err == nil && len(payloadBytes) > 0 {
		values, _ := url.ParseQuery(string(payloadBytes))
		role = values.Get("role")
		accountType = values.Get("account_type")
	}
	if role == "" {
		role = "car_owner"
	}
	if accountType == "" {
		accountType = "car_owner"
	}

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

	userRepo := repositories.NewUserRepository(database.DB)
	existingUser, err := userRepo.FindUserByEmail(gi.Email)
	var u *models.User
	now := time.Now()
	if err == nil && existingUser != nil {
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logrus.Error("missing JWT_SECRET env")
		utils.Error(w, http.StatusInternalServerError, "server misconfigured")
		return
	}
	claims := jwt.MapClaims{
		"user_id":      u.ID.Hex(),
		"email":        u.Email,
		"role":         u.Role,
		"account_type": u.AccountType,
		"exp":          time.Now().Add(72 * time.Hour).Unix(),
	}
	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := tokenJwt.SignedString([]byte(jwtSecret))
	if err != nil {
		logrus.Error("failed to sign jwt: ", err)
		utils.Error(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    signedToken,
		Path:     "/",
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Always true for SameSite=None
		SameSite: http.SameSiteNoneMode,
	})

	// Redirect to frontend callback
	frontendURL := os.Getenv("FRONTEND_URL")
	callbackPath := "/CallbackPage?token=" + signedToken // Pass token in URL as fallback/initial
	redirectURL := fmt.Sprintf("%s%s", frontendURL, callbackPath)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// ForgotPasswordHandler - Send password reset email
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if payload.Email == "" {
		utils.Error(w, http.StatusBadRequest, "Email is required")
		return
	}

	userRepo := repositories.NewUserRepository(database.DB)
	user, err := userRepo.FindUserByEmail(payload.Email)

	// Don't reveal if email exists (security best practice)
	if err != nil {
		logrus.Infof("Password reset requested for non-existent email: %s", payload.Email)
		utils.JSON(w, http.StatusOK, map[string]string{
			"message": "If the email exists, a reset link has been sent",
		})
		return
	}

	// Generate 6-digit reset token
	resetToken, err := utils.GenerateNumericCode(6)
	if err != nil {
		logrus.Error("Failed to generate reset token: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to process reset request")
		return
	}
	expiry := time.Now().Add(1 * time.Hour)

	// Save token to database
	err = userRepo.UpdateUserByID(user.ID, bson.M{
		"password_reset_token":  resetToken,
		"password_reset_expiry": expiry,
	})

	if err != nil {
		logrus.Error("Failed to save reset token: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to process reset request")
		return
	}

	// Send reset email
	err = utils.SendPasswordResetEmail(user.Email, user.Name, resetToken)
	if err != nil {
		logrus.Error("Failed to send reset email: ", err)
		// Don't fail the request if email fails, token is still saved
	}

	logrus.Infof("Password reset token generated for user: %s", user.Email)
	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a reset link has been sent",
	})
}

// ResetPasswordHandler - Reset password with token
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate inputs
	if payload.Token == "" || payload.NewPassword == "" {
		utils.Error(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Validate password length
	if len(payload.NewPassword) < 6 {
		utils.Error(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	userRepo := repositories.NewUserRepository(database.DB)
	user, err := userRepo.FindUserByResetToken(payload.Token)

	if err != nil {
		logrus.Warn("Invalid or expired reset token")
		utils.Error(w, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(payload.NewPassword)
	if err != nil {
		logrus.Error("Failed to hash password: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	// Update password and clear reset token
	err = userRepo.UpdateUserByID(user.ID, bson.M{
		"password":              hashedPassword,
		"password_reset_token":  "",
		"password_reset_expiry": time.Time{},
		"updated_at":            time.Now(),
	})

	if err != nil {
		logrus.Error("Failed to update password: ", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	logrus.Infof("Password successfully reset for user: %s", user.Email)
	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successful",
	})
}

func (ac *AuthController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteStrictMode,
	})
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
