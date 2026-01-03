package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/olabanji12-ojo/CarWashApp/controllers"
	// "github.com/olabanji12-ojo/CarWashApp/middleware"
)

func AuthRoutes(router *mux.Router, authService *controllers.AuthController) {

	// Base route to confirm it's working
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(" API is live!"))
	}).Methods("GET")

	// Group: /api/auth the authentication
	auth := router.PathPrefix("/api/auth").Subrouter()

	// auth.Use(middleware.AuthMiddleware)

	// POST /api/auth/register
	auth.HandleFunc("/register", authService.RegisterHandler).Methods("POST")

	// POST /api/auth/login
	auth.HandleFunc("/login", authService.LoginHandler).Methods("POST")

	// GET /api/auth/google/login
	auth.HandleFunc("/google/login", authService.GoogleLoginHandler).Methods("GET")

	// GET /api/auth/google/callback
	auth.HandleFunc("/google/callback", authService.GoogleCallbackHandler).Methods("GET")

	// POST /api/auth/logout
	auth.HandleFunc("/logout", authService.LogoutHandler).Methods("POST")

	// POST /api/auth/verify
	auth.HandleFunc("/verify", authService.VerifyEmailHandler).Methods("POST")

	// POST /api/auth/resend-verification
	auth.HandleFunc("/resend-verification", authService.ResendVerificationEmailHandler).Methods("POST")

	// POST /api/auth/forgot-password
	auth.HandleFunc("/forgot-password", controllers.ForgotPasswordHandler).Methods("POST")

	// POST /api/auth/reset-password
	auth.HandleFunc("/reset-password", controllers.ResetPasswordHandler).Methods("POST")

}