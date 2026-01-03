package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

// WorkerRouter handles worker-related routing
type WorkerRouter struct {
	workerController *controllers.WorkerController
}

// NewWorkerRouter creates a new WorkerRouter instance
func NewWorkerRouter(workerController *controllers.WorkerController) *WorkerRouter {
	return &WorkerRouter{workerController: workerController}
}

// WorkerRoutes sets up all worker-related routes
func (wr *WorkerRouter) WorkerRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api/workers").Subrouter()

	// Apply auth middleware to all worker routes
	subRouter.Use(middleware.AuthMiddleware)

	// Worker CRUD operations
	subRouter.HandleFunc("/create", wr.workerController.CreateWorkersForBusiness).Methods("POST")
	subRouter.HandleFunc("/business/{id}", wr.workerController.GetWorkersForBusiness).Methods("GET")

	// Worker status management
	subRouter.HandleFunc("/status/{id}", wr.workerController.UpdateWorkerStatus).Methods("PATCH")
	subRouter.HandleFunc("/work-status/{id}", wr.workerController.UpdateWorkerWorkStatus).Methods("PATCH")

	// Worker assignment functionality
	subRouter.HandleFunc("/available/{id}", wr.workerController.GetAvailableWorkersForBusiness).Methods("GET")
	subRouter.HandleFunc("/assign", wr.workerController.AssignWorkerToOrder).Methods("POST")
	subRouter.HandleFunc("/remove", wr.workerController.RemoveWorkerFromOrder).Methods("POST")

	// Worker profile management
	subRouter.HandleFunc("/{id}", wr.workerController.UpdateWorkerHandler).Methods("PUT", "OPTIONS")
	subRouter.HandleFunc("/{id}/photo", wr.workerController.UploadWorkerPhotoHandler).Methods("POST", "OPTIONS")
}
