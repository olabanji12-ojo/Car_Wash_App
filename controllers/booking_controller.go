package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"

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

	if err := input.Validate(); err != nil {
	utils.Error(w, http.StatusBadRequest, err.Error())
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
	dateStr := r.URL.Query().Get("date") // ?date=2025-07-08 or ?date=2025-07-08T14:00:00Z
	logrus.Info("Recieved date query param:", dateStr)

	var date time.Time
	var err error

	if dateStr == "" {
	utils.Error(w, http.StatusBadRequest, "Missing date parameter (?date=YYYY-MM-DD)")
	return
}


	// Try full datetime first 
	date, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// If that fails, try just YYYY-MM-DD
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD or RFC3339)")
			return
		}
	}

	bookings, err := services.GetBookingsByDate(carwashID, date)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
}


func UpdateBookingHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID
	bookingID := mux.Vars(r)["bookingID"]
	logrus.Info("Vars:", bookingID)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update data")
		return
	}

	err := services.UpdateBooking(userID, bookingID, updates)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Car updated successfully"})
}

