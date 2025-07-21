package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

func BookingRoutes(router *mux.Router) {

	booking := router.PathPrefix("/api/bookings").Subrouter()

	//  JWT Middleware for all protected booking routes
	booking.Use(middleware.AuthMiddleware)
    
	// POST /api/bookings
	booking.HandleFunc("", controllers.CreateBookingHandler).Methods("POST") // tested 
    
	// GET /api/bookings/{id}
	booking.HandleFunc("/{id}", controllers.GetBookingByIDHandler).Methods("GET") // tested 
    
	// GET /api/bookings/user/me
	booking.HandleFunc("/user/me", controllers.GetMyBookingsHandler).Methods("GET") // tested 
    
	// GET /api/bookings/carwash/{carwash_id}
	booking.HandleFunc("/carwash/{carwash_id}", controllers.GetBookingsByCarwashHandler).Methods("GET") // tested 

    // PUT/api/bookings/{id}
    booking.HandleFunc("/booking/{bookingID}", controllers.UpdateBookingHandler).Methods("PUT")
   
	// PUT /api/bookings/{id}/status
	booking.HandleFunc("/{id}/status", controllers.UpdateBookingStatusHandler).Methods("PUT") // tested 

	// DELETE /api/bookings/{id}
	booking.HandleFunc("/{id}", controllers.CancelBookingHandler).Methods("DELETE") // tested 

	// Optional: GET /api/bookings/carwash/{carwash_id}/date?date=2025-07-08
	booking.HandleFunc("/carwash/{carwash_id}/date", controllers.GetBookingsByDateHandler).Methods("GET") // tested

	
}



