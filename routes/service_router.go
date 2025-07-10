package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

// ServiceRoutes sets up all routes for service-related endpoints
func ServiceRoutes(router *mux.Router) {
	serviceRouter := router.PathPrefix("/api/services").Subrouter()

	// Only authenticated users can access these routes
	serviceRouter.Use(middleware.AuthMiddleware)

	// POST /api/services        -> Create a service (business only)
	serviceRouter.HandleFunc("", controllers.CreateServiceHandler).Methods("POST")

	// GET /api/services/my      -> Get all services for current business
	serviceRouter.HandleFunc("/my", controllers.GetMyServicesHandler).Methods("GET")

	// GET /api/services/{id}    -> Get one service by ID (publicly accessible)
	router.HandleFunc("/api/services/{id}", controllers.GetServiceByIDHandler).Methods("GET")

	// PUT /api/services/{id}    -> Update a service (business only)
	serviceRouter.HandleFunc("/{id}", controllers.UpdateServiceHandler).Methods("PUT")

	// DELETE /api/services/{id} -> Soft delete service (business only)
	serviceRouter.HandleFunc("/{id}", controllers.DeleteServiceHandler).Methods("DELETE")
}