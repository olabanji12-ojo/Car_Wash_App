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

func (cwr *CarWashRouter) CarwashRoutes(router *mux.Router) {
	carwash := router.PathPrefix("/api/carwashes").Subrouter()
	carwash.Use(middleware.AuthMiddleware)

	// Public routes
	carwash.HandleFunc("", cwr.carwashController.GetAllActiveCarwashesHandler).Methods("GET")
	carwash.HandleFunc("/{id}", cwr.carwashController.GetCarwashByIDHandler).Methods("GET")
	carwash.HandleFunc("/nearby/", cwr.carwashController.GetNearbyCarwashesHandler).Methods("GET")


	// Owner-specific routes (authenticated)
	carwash.HandleFunc("", cwr.carwashController.CreateCarwashHandler).Methods("POST")
	carwash.HandleFunc("/{id}", cwr.carwashController.UpdateCarwashHandler).Methods("PUT")
	carwash.HandleFunc("/{id}/status", cwr.carwashController.SetCarwashStatusHandler).Methods("PATCH")
	carwash.HandleFunc("/{id}/location", cwr.carwashController.UpdateCarwashLocationHandler).Methods("PUT")
	carwash.HandleFunc("/owner/{owner_id}", cwr.carwashController.GetCarwashesByOwnerIDHandler).Methods("GET")
	carwash.HandleFunc("/carwash/{id}/onboarding", cwr.carwashController.CompleteOnboarding).Methods("PUT")

	// Service management routes
	carwash.HandleFunc("/services/{carwashid}", cwr.carwashController.CreateServiceHandler).Methods("POST")
	carwash.HandleFunc("/services/carwash/{carwashid}", cwr.carwashController.GetServicesHandler).Methods("GET")

	

	carwash.HandleFunc("/services/{carwashid}", cwr.carwashController.UpdateServiceHandler).Methods("PUT")
	carwash.HandleFunc("/services/{carwashid}", cwr.carwashController.DeleteServiceHandler).Methods("DELETE")
}


