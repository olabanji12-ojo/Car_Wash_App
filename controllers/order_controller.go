package controllers

import (

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
)


type OrderController struct {
	OrderService *services.OrderService
}

func NewOrderController(orderService *services.OrderService) *OrderController {
	return &OrderController{OrderService: orderService}
}

//  Create Order from Booking (business only)
func(oc *OrderController) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	
	ctx := r.Context().Value("auth").(middleware.AuthContext)
	// authCtx := r.Context().Value("auth").(middleware.AuthContext)
	
	if ctx.Role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can create orders")
		return
	}

	bookingID := mux.Vars(r)["booking_id"]
	
	order, err := oc.OrderService.CreateOrderFromBooking(bookingID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, order)
}

//  Get Order by ID (owner or business)
func(oc *OrderController) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["order_id"]

	order, err := oc.OrderService.GetOrderByID(orderID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, order)
}

//  Get all orders for logged-in user (car owner)

func(oc *OrderController) GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("auth").(middleware.AuthContext)

	orders, err := oc.OrderService.GetOrdersByUser(ctx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, orders)

}


//  Get all orders for carwash (business)
func(oc *OrderController) GetCarwashOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("auth").(middleware.AuthContext)

	if ctx.Role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can view their orders")
		return
	}

	orders, err := oc.OrderService.GetOrdersByCarwash(ctx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, orders)
}

//  Update Order Status (business)
func(oc *OrderController) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("auth").(middleware.AuthContext)

	if ctx.Role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can update order status")
		return
	}

	orderID := mux.Vars(r)["order_id"]

	var input struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Status == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid status update")
		return
	}

	if err := oc.OrderService.UpdateOrderStatus(orderID, input.Status); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Order status updated"})
}

//  Assign Worker to Order (optional)
func(oc *OrderController) AssignWorkerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("auth").(middleware.AuthContext)

	if ctx.Role != "business" {
		utils.Error(w, http.StatusForbidden, "Only businesses can assign workers")
		return
	}

	orderID := mux.Vars(r)["order_id"]

	var input struct {
		WorkerID string `json:"worker_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.WorkerID == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid worker ID")
		return
	}

	if err := oc.OrderService.AssignWorker(orderID, input.WorkerID); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Worker assigned to order"})
}



