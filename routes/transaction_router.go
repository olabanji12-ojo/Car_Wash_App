package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

//  Register payment-related routes here
func PaymentRoutes(router *mux.Router) {


	payment := router.PathPrefix("/api/payments").Subrouter()
	payment.Use(middleware.AuthMiddleware) // Protect all routes

	payment.HandleFunc("", controllers.CreatePaymentHandler).Methods("POST") // tested
	payment.HandleFunc("/payment/{id}", controllers.GetPaymentByIDHandler).Methods("GET") // testing
	payment.HandleFunc("/user", controllers.GetPaymentsByUserHandler).Methods("GET") // tested
	payment.HandleFunc("/carwash/{id}", controllers.GetPaymentsByCarwashHandler).Methods("GET") // tested

}


