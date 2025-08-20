package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)


type CarWashRouter struct {
	carwashController controllers.CarWashController
}

func NewCarWashRouter(carwashController controllers.CarWashController) *CarWashRouter {
	return &CarWashRouter{carwashController: carwashController}
}


// CarwashRoutes sets up all routes for carwash-related actions
func(cwr *CarWashRouter) CarwashRoutes(router *mux.Router) {
      
	carwash := router.PathPrefix("/api/carwashes").Subrouter()
    
	carwash.Use(middleware.AuthMiddleware)

	// Public routes
	carwash.HandleFunc("", cwr.carwashController.GetAllActiveCarwashesHandler).Methods("GET") // tested
	carwash.HandleFunc("/{id}", cwr.carwashController.GetCarwashByIDHandler).Methods("GET")  //  tested 

	carwash.HandleFunc("/nearby", cwr.carwashController.GetNearbyCarwashesHandler).Methods("GET") // location-based search
    
	// Owner-specific routes (authenticated)
	carwash.HandleFunc("", cwr.carwashController.CreateCarwashHandler).Methods("POST") // tested
	carwash.HandleFunc("/{id}", cwr.carwashController.UpdateCarwashHandler).Methods("PUT") // tested
	carwash.HandleFunc("/{id}/status", cwr.carwashController.SetCarwashStatusHandler).Methods("PATCH") // tested
	carwash.HandleFunc("/{id}/location", cwr.carwashController.UpdateCarwashLocationHandler).Methods("PUT") // location update
	carwash.HandleFunc("/owner/{owner_id}", cwr.carwashController.GetCarwashesByOwnerIDHandler).Methods("GET") // tested 
	
    
}

