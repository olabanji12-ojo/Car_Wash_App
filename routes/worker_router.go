package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)


func WorkerRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api/workers").Subrouter()

	subRouter.Use(middleware.AuthMiddleware)

	// Existing routes
	subRouter.HandleFunc("/create", controllers.CreateWorkersForBusiness).Methods("POST")

	subRouter.HandleFunc("/business/{id}", controllers.GetWorkersForBusiness).Methods("GET")
	subRouter.HandleFunc("/status/{id}", controllers.UpdateWorkerStatus).Methods("PATCH")

	// New routes for worker assignment functionality
	subRouter.HandleFunc("/available/{id}", controllers.GetAvailableWorkersForBusiness).Methods("GET")
	subRouter.HandleFunc("/work-status/{id}", controllers.UpdateWorkerWorkStatus).Methods("PATCH")
	subRouter.HandleFunc("/assign", controllers.AssignWorkerToOrder).Methods("POST")
	subRouter.HandleFunc("/remove", controllers.RemoveWorkerFromOrder).Methods("POST")

	
}

