package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

func OrderRoutes(router *mux.Router) {


	// Prefix: /api/orders

	orderRouter := router.PathPrefix("/api/orders").Subrouter()
	orderRouter.Use(middleware.AuthMiddleware)

	//  Create order from approved booking
	
	orderRouter.HandleFunc("/booking/{booking_id}", controllers.CreateOrderHandler).Methods("POST") // tested
	//  Get specific order
	orderRouter.HandleFunc("/order/{order_id}", controllers.GetOrderByIDHandler).Methods("GET") // tested 

	//  Get logged-in user's orders (car owner)
	orderRouter.HandleFunc("/my", controllers.GetUserOrdersHandler).Methods("GET") // tested

	//  Get business orders (business user)
	orderRouter.HandleFunc("/business", controllers.GetCarwashOrdersHandler).Methods("GET") // tested

	//  Update order status (e.g. completed, in_progress)
	orderRouter.HandleFunc("/{order_id}/status", controllers.UpdateOrderStatusHandler).Methods("PATCH") // tested

	//  Assign a worker (optional)
	orderRouter.HandleFunc("/{order_id}/assign", controllers.AssignWorkerHandler).Methods("PATCH") // to be built later


}



