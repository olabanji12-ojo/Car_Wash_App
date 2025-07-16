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
	carwash.HandleFunc("", controllers.GetAllActiveCarwashesHandler).Methods("GET") // tested
	carwash.HandleFunc("/{id}", controllers.GetCarwashByIDHandler).Methods("GET")  //  tested

	// Owner-specific routes (authenticated)
	carwash.HandleFunc("", controllers.CreateCarwashHandler).Methods("POST") // tested
	carwash.HandleFunc("/{id}", controllers.UpdateCarwashHandler).Methods("PUT") // tested
	carwash.HandleFunc("/{id}/status", controllers.SetCarwashStatusHandler).Methods("PATCH") // tested
	carwash.HandleFunc("/owner/{owner_id}", controllers.GetCarwashesByOwnerIDHandler).Methods("GET") // tested 
	
    
}



