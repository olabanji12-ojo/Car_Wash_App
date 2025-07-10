package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/olabanji12-ojo/CarWashApp/controllers"

	"github.com/olabanji12-ojo/CarWashApp/middleware"
)




func AuthRoutes(router *mux.Router)  {
	
	// Base route to confirm it's working
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🚀 API is live!"))
	}).Methods("GET")

	// Group: /api/auth the authentication
	auth := router.PathPrefix("/api/auth").Subrouter()

	auth.Use(middleware.AuthMiddleware)

	// POST /api/auth/register
	auth.HandleFunc("/register", controllers.RegisterHandler).Methods("POST")

	// POST /api/auth/login
	auth.HandleFunc("/login", controllers.LoginHandler).Methods("POST")




}
