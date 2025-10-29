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

func (br *BookingRouter) BookingRoutes(router *mux.Router) {
	// Public routes (no auth)
	publicBooking := router.PathPrefix("/api/bookings").Subrouter()
	publicBooking.HandleFunc("/carwash/{carwash_id}/slots", br.bookingController.GetAvailableSlotsHandler).Methods("GET")

	// Protected routes (require auth)
	protectedBooking := router.PathPrefix("/api/bookings").Subrouter()
	protectedBooking.Use(middleware.AuthMiddleware)

	// POST /api/bookings
	protectedBooking.HandleFunc("", br.bookingController.CreateBookingHandler).Methods("POST")

	// GET /api/bookings/{id}
	protectedBooking.HandleFunc("/{id}", br.bookingController.GetBookingByIDHandler).Methods("GET")

	// GET /api/bookings/user/me
	protectedBooking.HandleFunc("/user/me", br.bookingController.GetMyBookingsHandler).Methods("GET")

	// GET /api/bookings/carwash/{carwash_id}
	protectedBooking.HandleFunc("/carwash/{carwash_id}", br.bookingController.GetBookingsByCarwashHandler).Methods("GET")

	// PUT /api/bookings/{id}
	protectedBooking.HandleFunc("/{id}", br.bookingController.UpdateBookingHandler).Methods("PUT")

	// PATCH /api/bookings/{id}/status
	protectedBooking.HandleFunc("/{id}/status", br.bookingController.UpdateBookingStatusHandler).Methods("PATCH")

	// DELETE /api/bookings/{id}
	protectedBooking.HandleFunc("/{id}", br.bookingController.CancelBookingHandler).Methods("DELETE")

	// GET /api/bookings/carwash/{carwash_id}/date?date=YYYY-MM-DD
	protectedBooking.HandleFunc("/carwash/{carwash_id}/date", br.bookingController.GetBookingsByDateHandler).Methods("GET")

	// GET /api/bookings/carwash/{carwash_id}/filter
	protectedBooking.HandleFunc("/carwash/{carwash_id}/filter", br.bookingController.GetBookingsByCarwashWithFiltersHandler).Methods("GET")
}