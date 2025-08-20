package routes


import (

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	// "github.com/olabanji12-ojo/CarWashApp/services"

)


type OrderRouter struct {
	orderController *controllers.OrderController
}

func NewOrderRouter(orderController *controllers.OrderController) *OrderRouter {
	return &OrderRouter{orderController: orderController}
}



func(or *OrderRouter) OrderRoutes(router *mux.Router) {
    
	// Prefix: /api/orders

	orderRouter := router.PathPrefix("/api/orders").Subrouter()
	orderRouter.Use(middleware.AuthMiddleware)

	//  Create order from approved booking
	
	orderRouter.HandleFunc("/booking/{booking_id}", or.orderController.CreateOrderHandler).Methods("POST") // tested
	//  Get specific order
	orderRouter.HandleFunc("/order/{order_id}", or.orderController.GetOrderByIDHandler).Methods("GET") // tested 

	//  Get logged-in user's orders (car owner)
	orderRouter.HandleFunc("/my", or.orderController.GetUserOrdersHandler).Methods("GET") // tested

	//  Get business orders (business user)
	orderRouter.HandleFunc("/business", or.orderController.GetCarwashOrdersHandler).Methods("GET") // tested

	//  Update order status (e.g. completed, in_progress)
	orderRouter.HandleFunc("/{order_id}/status", or.orderController.UpdateOrderStatusHandler).Methods("PATCH") // tested

	//  Assign a worker (optional)
	orderRouter.HandleFunc("/{order_id}/assign", or.orderController.AssignWorkerHandler).Methods("PATCH") // to be built later


}



