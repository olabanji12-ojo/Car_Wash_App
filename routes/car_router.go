package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)


type CarRouter struct {
	carController controllers.CarController
}


func NewCarRouter(carController controllers.CarController) *CarRouter {
	return &CarRouter{carController: carController}
}

func (cr *CarRouter)CarRoutes(router *mux.Router) {
	car := router.PathPrefix("/api/cars").Subrouter()

	// Apply auth middleware to all car routes 
    
	car.Use(middleware.AuthMiddleware)
    
	car.HandleFunc("/", cr.carController.CreateCarHandler).Methods("POST") //        
	car.HandleFunc("/my", cr.carController.GetMyCarsHandler).Methods("GET") // tested      
	car.HandleFunc("/{carID}", cr.carController.GetCarByIDHandler).Methods("GET") // tested
    
	car.HandleFunc("/update/{carID}", cr.carController.UpdateCarHandler).Methods("PUT") // tested 
	car.HandleFunc("/{carID}", cr.carController.DeleteCarHandler).Methods("DELETE") //  

	car.HandleFunc("/{carID}/default", cr.carController.SetDefaultCarHandler).Methods("PATCH") // tested
    
	
}


