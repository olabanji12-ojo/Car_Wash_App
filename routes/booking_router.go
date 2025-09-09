package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)


type BookingRouter struct {
	bookingController controllers.BookingController
}

func NewBookingRouter(bookingController controllers.BookingController) *BookingRouter {

    return &BookingRouter{bookingController: bookingController}

}
 

func(br *BookingRouter) BookingRoutes(router *mux.Router) {

	booking := router.PathPrefix("/api/bookings").Subrouter()

	//  JWT Middleware for all protected booking routes
	booking.Use(middleware.AuthMiddleware)
    
	// POST /api/bookings
	booking.HandleFunc("", br.bookingController.CreateBookingHandler).Methods("POST") // tested 
    
	// GET /api/bookings/{id}
	booking.HandleFunc("/{id}", br.bookingController.GetBookingByIDHandler).Methods("GET") // tested 
    
	// GET /api/bookings/user/me
	booking.HandleFunc("/user/me", br.bookingController.GetMyBookingsHandler).Methods("GET") // tested 
    
	// GET /api/bookings/carwash/{carwash_id}
	booking.HandleFunc("/carwash/{carwash_id}", br.bookingController.GetBookingsByCarwashHandler).Methods("GET") // tested 

    // PUT/api/bookings/{id}
    booking.HandleFunc("/booking/{bookingID}", br.bookingController.UpdateBookingHandler).Methods("PUT")
   
	// PATCH /api/bookings/{id}/status
	booking.HandleFunc("/{id}/status", br.bookingController.UpdateBookingStatusHandler).Methods("PATCH") // tested 

	// DELETE /api/bookings/{id}
	booking.HandleFunc("/{id}", br.bookingController.CancelBookingHandler).Methods("DELETE") // tested 

	// Optional: GET /api/bookings/carwash/{carwash_id}/date?date=2025-07-08
	booking.HandleFunc("/carwash/{carwash_id}/date", br.bookingController.GetBookingsByDateHandler).Methods("GET") // tested

	booking.HandleFunc("/carwash/{carwash_id}/filter", br.bookingController.GetBookingsByCarwashWithFiltersHandler).Methods("GET")
    
	
}

