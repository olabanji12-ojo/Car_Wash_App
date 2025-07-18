package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
)

func WorkerRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api/workers").Subrouter()

	subRouter.HandleFunc("/business/{id}", controllers.GetWorkersForBusiness).Methods("GET")
	subRouter.HandleFunc("/status/{id}", controllers.UpdateWorkerStatus).Methods("PATCH")
}
