package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

func CarRoutes(router *mux.Router) {
	car := router.PathPrefix("/api/cars").Subrouter()

	// Apply auth middleware to all car routes
	car.Use(middleware.AuthMiddleware)

	car.HandleFunc("/", controllers.CreateCarHandler).Methods("POST")
	car.HandleFunc("/my", controllers.GetMyCarsHandler).Methods("GET")
	car.HandleFunc("/{carID}", controllers.UpdateCarHandler).Methods("PUT")
	car.HandleFunc("/{carID}", controllers.DeleteCarHandler).Methods("DELETE")
	car.HandleFunc("/{carID}/default", controllers.SetDefaultCarHandler).Methods("PATCH")
}
