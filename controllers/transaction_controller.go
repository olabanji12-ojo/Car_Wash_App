
package controllers

import (
	
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

)

//  POST /api/payments → Create a new payment
func CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	var input models.Payment
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := input.Validate(); err != nil {
	utils.Error(w, http.StatusBadRequest, err.Error())
	return
   }

	created, err := services.CreatePayment(userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, created)
}

//  GET /api/payments/{id} → Get payment by ID
func GetPaymentByIDHandler(w http.ResponseWriter, r *http.Request) {
	paymentID := mux.Vars(r)["id"]

	payment, err := services.GetPaymentByOrderID(paymentID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}
    
	utils.JSON(w, http.StatusOK, payment)
	
}

// GET /api/payments/user → Get all user payments
func GetPaymentsByUserHandler(w http.ResponseWriter, r *http.Request) {

	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID  := authCtx.UserID

	payments, err := services.GetPaymentsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, payments)
}

//  GET /api/payments/carwash/{id} → All for a carwash
func GetPaymentsByCarwashHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["id"]

	payments, err := services.GetPaymentsByCarwashID(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, payments)
}


