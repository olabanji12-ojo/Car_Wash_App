package controllers

import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"

	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingController struct {
	BookingService *services.BookingService
	CarWashService *services.CarWashService
}

func NewBookingController(bookingService *services.BookingService, carWashService *services.CarWashService) *BookingController {
	return &BookingController{
		BookingService: bookingService,
		CarWashService: carWashService,
	}
}

// 1. CreateBookingHandler
func (bc *BookingController) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {

	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID
	role := authCtx.Role
	accountType := authCtx.AccountType

	var input models.Booking
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid booking data")
		return
	}

	// Status validation check - only allow specific values
	validStatuses := []string{"pending", "confirmed", "completed", "cancelled"}
	if input.Status != "" {
		isValidStatus := false
		for _, validStatus := range validStatuses {
			if input.Status == validStatus {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			utils.Error(w, http.StatusBadRequest, "Invalid status. Must be one of: pending, confirmed, completed, cancelled")
			return
		}
	} else {
		// If status is empty, set default to pending
		input.Status = "pending"
	}

	// Now validate the rest of the booking
	if err := input.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if role != "car_owner" || accountType != "car_owner" {
		utils.Error(w, http.StatusForbidden, "Only car owners can create bookings")
		return
	}

	newBooking, err := bc.BookingService.CreateBooking(userID, input)
	if err != nil {
		// Log the actual error for server-side debugging
		logrus.Errorf("[CreateBooking] Failed to create booking: %v", err)

		// Map specific business logic errors to 400 Bad Request
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "outside the delivery radius") ||
			strings.Contains(errorMessage, "busy") ||
			strings.Contains(errorMessage, "already have a booking") ||
			strings.Contains(errorMessage, "outside of open hours") ||
			strings.Contains(errorMessage, "fully booked") ||
			strings.Contains(errorMessage, "location coordinates are required") {
			utils.Error(w, http.StatusBadRequest, errorMessage)
			return
		}

		utils.Error(w, http.StatusInternalServerError, "An unexpected server error occurred: "+errorMessage)
		return
	}

	utils.JSON(w, http.StatusCreated, newBooking)
}

// 2. GetBookingByIDHandler
func (bc *BookingController) GetBookingByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	booking, err := bc.BookingService.GetBookingByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, booking)
}

// 3. GetMyBookingsHandler
func (bc *BookingController) GetMyBookingsHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	bookings, err := bc.BookingService.GetBookingsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
}

// 4. GetBookingsByCarwashHandler
func (bc *BookingController) GetBookingsByCarwashHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]

	bookings, err := bc.BookingService.GetBookingsByCarwashID(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, bookings)
}

// 5. UpdateBookingStatusHandler
func (bc *BookingController) UpdateBookingStatusHandler(w http.ResponseWriter, r *http.Request) {
	bookingID := mux.Vars(r)["id"]

	var payload struct {
		Status           string `json:"status"`
		VerificationCode string `json:"verification_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Status == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid status input")
		return
	}

	if err := bc.BookingService.UpdateBookingStatus(bookingID, payload.Status, payload.VerificationCode); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Booking status updated"})
}

// 6. CancelBookingHandler
func (bc *BookingController) CancelBookingHandler(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := bc.BookingService.CancelBooking(id); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Booking cancelled"})

}

// 7. GetBookingsByDateHandler (Optional)
func (bc *BookingController) GetBookingsByDateHandler(w http.ResponseWriter, r *http.Request) {
	carwashIDStr := mux.Vars(r)["carwash_id"]
	dateStr := r.URL.Query().Get("date")

	// COMPREHENSIVE DEBUG LOGGING
	logrus.Info("🔍 INCOMING REQUEST DEBUG:")
	logrus.Info("Raw carwashID from URL:", carwashIDStr)
	logrus.Info("Raw dateStr from query:", dateStr)
	logrus.Info("Full request URL:", r.URL.String())
	logrus.Info("Request method:", r.Method)
	logrus.Info("Authorization header present:", r.Header.Get("Authorization") != "")
	logrus.Info("carwashIDStr type:", fmt.Sprintf("%T", carwashIDStr))
	logrus.Info("carwashIDStr length:", len(carwashIDStr))

	var date time.Time
	var err error

	if dateStr == "" {
		logrus.Error("❌ Missing date parameter")
		utils.Error(w, http.StatusBadRequest, "Missing date parameter (?date=YYYY-MM-DD)")
		return
	}

	// Parse carwash_id to ObjectID
	logrus.Info("🔄 CONVERTING CARWASH_ID:")
	logrus.Info("Attempting to convert carwashIDStr to ObjectID:", carwashIDStr)

	carwashID, err := primitive.ObjectIDFromHex(carwashIDStr)
	if err != nil {
		logrus.Error("❌ Invalid carwash_id format - cannot convert to ObjectID:", err)
		logrus.Error("carwashIDStr value:", carwashIDStr)
		utils.Error(w, http.StatusBadRequest, "Invalid carwash_id format")
		return
	}

	logrus.Info("✅ Successfully converted to ObjectID:", carwashID)
	logrus.Info("ObjectID hex representation:", carwashID.Hex())

	// Parse date
	logrus.Info("🔄 PARSING DATE:")
	logrus.Info("Attempting to parse dateStr:", dateStr)

	// Try full datetime first
	date, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		logrus.Info("⚠️ RFC3339 parse failed, trying YYYY-MM-DD format")
		logrus.Info("RFC3339 error:", err)

		// If that fails, try just YYYY-MM-DD
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			logrus.Error("❌ Date parsing failed for both formats:", err)
			utils.Error(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD or RFC3339)")
			return
		}
	}

	logrus.Info("✅ SUCCESSFULLY PARSED:")
	logrus.Info("Parsed date:", date)
	logrus.Info("Parsed carwashID:", carwashID)
	logrus.Info("Date in UTC:", date.UTC())
	logrus.Info("Date format check:", date.Format("2006-01-02"))
	logrus.Info("Date location:", date.Location())

	// Call the service
	logrus.Info("🚀 CALLING BookingService.GetBookingsByDate")
	logrus.Info("Parameters - carwashID:", carwashID, "date:", date)

	bookings, err := bc.BookingService.GetBookingsByDate(carwashIDStr, date)
	if err != nil {
		logrus.Error("❌ BookingService.GetBookingsByDate error:", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.Info("✅ BOOKINGS RESULT:")
	logrus.Info("Bookings found count:", len(bookings))
	logrus.Info("Bookings data:", bookings)

	// Log each booking individually for detailed inspection
	for i, booking := range bookings {
		logrus.Info(fmt.Sprintf("📋 BOOKING %d:", i))
		logrus.Info("  ID:", booking.ID)
		logrus.Info("  CarwashID:", booking.CarwashID)
		logrus.Info("  BookingTime:", booking.BookingTime)
		logrus.Info("  Status:", booking.Status)
		// Add other relevant fields as needed
	}

	logrus.Info("📤 SENDING RESPONSE:")
	logrus.Info("Response will contain", len(bookings), "bookings")

	utils.JSON(w, http.StatusOK, bookings)
}

func (bc *BookingController) UpdateBookingHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID
	bookingID := mux.Vars(r)["bookingID"]
	logrus.Info("Vars:", bookingID)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid update data")
		return
	}

	err := bc.BookingService.UpdateBooking(userID, bookingID, updates)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Car updated successfully"})
}

// Get Bookings with filter from the controller

func (bc *BookingController) GetBookingsByCarwashWithFiltersHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]
	status := r.URL.Query().Get("status")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	logrus.Info("🔍 INCOMING REQUEST DEBUG:")
	logrus.Info("carwash_id:", carwashID)
	logrus.Info("status:", status)
	logrus.Info("from:", from)
	logrus.Info("to:", to)

	bookings, err := bc.BookingService.GetBookingsByCarwashWithFilters(carwashID, status, from, to)
	if err != nil {
		logrus.Error("❌ Failed to fetch bookings: ", err)
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Info("✅ BOOKINGS RESULT:")
	logrus.Info("Bookings found count:", len(bookings))

	utils.JSON(w, http.StatusOK, map[string]interface{}{"data": bookings})
}

func (bc *BookingController) GetAvailableSlotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carwashID := vars["carwash_id"]
	if carwashID == "" {
		logrus.Error("Missing carwash_id parameter")
		utils.Error(w, http.StatusBadRequest, "Missing carwash_id parameter")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		logrus.Error("Missing date parameter")
		utils.Error(w, http.StatusBadRequest, "Missing date parameter")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		logrus.Error("Invalid date format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Error("Invalid carwash ID format: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid carwash ID format")
		return
	}

	slots, err := bc.CarWashService.GetAvailableSlots(carwashObjID, date)
	if err != nil {
		logrus.Error("Failed to retrieve available slots: ", err)
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, slots)
}

// UpdateWorkerLocationHandler handles PATCH /api/bookings/:id/location
func (bc *BookingController) UpdateWorkerLocationHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var input struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid location data")
		return
	}

	if err := bc.BookingService.UpdateWorkerLocation(id, input.Lat, input.Lng); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Location updated"})
}

// GetPublicBookingHandler handles GET /api/bookings/track/:id
func (bc *BookingController) GetPublicBookingHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	booking, err := bc.BookingService.GetBookingByID(id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Booking not found")
		return
	}

	// Scrub sensitive data for public view
	publicInfo := map[string]interface{}{
		"id":            booking.ID,
		"status":        booking.Status,
		"booking_type":  booking.BookingType,
		"customer_name": booking.CustomerName,
		"notes":         booking.Notes,
		"address_note":  booking.AddressNote,
	}

	utils.JSON(w, http.StatusOK, publicInfo)
}
