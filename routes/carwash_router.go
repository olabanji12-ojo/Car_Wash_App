package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

// CarwashRoutes sets up all routes for carwash-related actions
func CarwashRoutes(router *mux.Router) {
	carwash := router.PathPrefix("/api/carwashes").Subrouter()

	carwash.Use(middleware.AuthMiddleware)

	// Public routes
	carwash.HandleFunc("", controllers.GetAllActiveCarwashesHandler).Methods("GET")
	carwash.HandleFunc("/{id}", controllers.GetCarwashByIDHandler).Methods("GET")

	// Owner-specific routes (authenticated)
	carwash.HandleFunc("", controllers.CreateCarwashHandler).Methods("POST")
	carwash.HandleFunc("/{id}", controllers.UpdateCarwashHandler).Methods("PUT")
	carwash.HandleFunc("/{id}/status", controllers.SetCarwashStatusHandler).Methods("PATCH")
	carwash.HandleFunc("/owner/{owner_id}", controllers.GetCarwashesByOwnerIDHandler).Methods("GET")
	
}