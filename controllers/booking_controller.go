package controllers

import (
    
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

	"time"
    
)

// 1. CreateBookingHandler
func CreateBookingHandler(w http.ResponseWriter, r *http.Request) {

	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	var input models.Booking
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid booking data")
		return
	}

	newBooking, err := services.CreateBooking(userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, newBooking)
}


// 2. GetBookingByIDHandler
func GetBookingByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	booking, err := services.GetBookingByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, booking)
}

// 3. GetMyBookingsHandler
func GetMyBookingsHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	bookings, err := services.GetBookingsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
}

// 4. GetBookingsByCarwashHandler
func GetBookingsByCarwashHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]

	bookings, err := services.GetBookingsByCarwashID(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
}

// 5. UpdateBookingStatusHandler
func UpdateBookingStatusHandler(w http.ResponseWriter, r *http.Request) {
	bookingID := mux.Vars(r)["id"]

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Status == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid status input")
		return
	}

	if err := services.UpdateBookingStatus(bookingID, payload.Status); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Booking status updated"})
}

// 6. CancelBookingHandler
func CancelBookingHandler(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := services.CancelBooking(id); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Booking cancelled"})

}

// 7. GetBookingsByDateHandler (Optional)
func GetBookingsByDateHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]
	dateStr := r.URL.Query().Get("date") // e.g. ?date=2025-07-08

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	bookings, err := services.GetBookingsByDate(carwashID, date)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
	
}