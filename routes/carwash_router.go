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
	// Public routes (no auth)
	publicCarwash := router.PathPrefix("/api/carwashes").Subrouter()
	publicCarwash.HandleFunc("", cwr.carwashController.GetAllActiveCarwashesHandler).Methods("GET")
	publicCarwash.HandleFunc("/nearby/", cwr.carwashController.GetNearbyCarwashesHandler).Methods("GET")
	publicCarwash.HandleFunc("/services/carwash/{carwashid}", cwr.carwashController.GetServicesHandler).Methods("GET")
	publicCarwash.HandleFunc("/services/{carwashid}/{serviceid}", cwr.carwashController.GetServiceByIDHandler).Methods("GET")

	// Protected routes (require auth)
	protectedCarwash := router.PathPrefix("/api/carwashes").Subrouter()
	protectedCarwash.Use(middleware.AuthMiddleware)

	// ✅ IMPORTANT: Register SPECIFIC routes BEFORE generic ones!
	// Gorilla Mux matches routes in order, so /{id}/photos must come before /{id}

	// Specific routes (with more path segments)
	protectedCarwash.HandleFunc("/{id}/photos", cwr.carwashController.UploadCarwashPhotoHandler).Methods("POST")
	protectedCarwash.HandleFunc("/{id}/status", cwr.carwashController.SetCarwashStatusHandler).Methods("PATCH")
	protectedCarwash.HandleFunc("/{id}/location", cwr.carwashController.UpdateCarwashLocationHandler).Methods("PUT")
	protectedCarwash.HandleFunc("/owner/{owner_id}", cwr.carwashController.GetCarwashesByOwnerIDHandler).Methods("GET")
	protectedCarwash.HandleFunc("/carwash/{id}/onboarding", cwr.carwashController.CompleteOnboarding).Methods("PUT")
	protectedCarwash.HandleFunc("/services/{carwashid}", cwr.carwashController.CreateServiceHandler).Methods("POST")
	protectedCarwash.HandleFunc("/services/{carwashid}/{serviceid}", cwr.carwashController.UpdateServiceHandler).Methods("PUT")
	protectedCarwash.HandleFunc("/services/{carwashid}/{serviceid}", cwr.carwashController.DeleteServiceHandler).Methods("DELETE")

	// Generic routes (less specific - MUST be last!)
	protectedCarwash.HandleFunc("", cwr.carwashController.CreateCarwashHandler).Methods("POST")
	protectedCarwash.HandleFunc("/{id}", cwr.carwashController.UpdateCarwashHandler).Methods("PUT")

	// Public GET /{id} - MUST be registered last to avoid conflicts!
	publicCarwash.HandleFunc("/{id}", cwr.carwashController.GetCarwashByIDHandler).Methods("GET")

}
