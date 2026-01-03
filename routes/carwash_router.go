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

func (cwr *CarWashRouter) CarwashRoutes(parentRouter *mux.Router) {
	carWashController := cwr.carwashController

	// Create a subrouter for /api/carwashes
	router := parentRouter.PathPrefix("/api/carwashes").Subrouter()

	// Public Routes (No Auth Required) - Allow Guest Browsing
	router.HandleFunc("/nearby", carWashController.GetNearbyCarwashesHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("", carWashController.GetAllActiveCarwashesHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/{carwashid}/services", carWashController.GetServicesHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/{carwashid}/services/{serviceid}", carWashController.GetServiceByIDHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/{id}", carWashController.GetCarwashByIDHandler).Methods("GET", "OPTIONS")

	// Protected Routes (Auth Required)
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("", carWashController.CreateCarwashHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/{id}", carWashController.UpdateCarwashHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/{id}/status", carWashController.SetCarwashStatusHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/{id}/complete-onboarding", carWashController.CompleteOnboarding).Methods("POST", "OPTIONS")
	protected.HandleFunc("/{id}/photos", carWashController.UploadCarwashPhotoHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/{carwashid}/services", carWashController.CreateServiceHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/{carwashid}/services/{serviceid}", carWashController.UpdateServiceHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/{carwashid}/services/{serviceid}", carWashController.DeleteServiceHandler).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/owner/{owner_id}", carWashController.GetCarwashesByOwnerIDHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/{id}/location", carWashController.UpdateCarwashLocationHandler).Methods("PUT", "OPTIONS")
}
