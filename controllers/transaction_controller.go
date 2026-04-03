package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionController struct {
	TransactionService *services.TransactionService
}

func NewTransactionController(ts *services.TransactionService) *TransactionController {
	return &TransactionController{
		TransactionService: ts,
	}
}

// GetEarningsSummaryHandler handles GET /api/transactions/summary/{carwash_id}
func (tc *TransactionController) GetEarningsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]
	if carwashID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing carwash ID")
		return
	}

	summary, err := tc.TransactionService.GetEarningsSummary(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve earnings summary")
		return
	}

	utils.JSON(w, http.StatusOK, summary)
}

// GetTransactionHistoryHandler handles GET /api/transactions/history/{carwash_id}
func (tc *TransactionController) GetTransactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["carwash_id"]
	if carwashID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing carwash ID")
		return
	}

	history, err := tc.TransactionService.GetTransactionHistory(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve transaction history")
		return
	}

	utils.JSON(w, http.StatusOK, history)
}

// GetPaymentsByUserHandler handles GET /api/payments/user
func (tc *TransactionController) GetPaymentsByUserHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	objID, _ := primitive.ObjectIDFromHex(userID)
	history, err := tc.TransactionService.GetTransactionHistoryByUserID(objID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve your payment history")
		return
	}

	utils.JSON(w, http.StatusOK, history)
}

// GetPaymentsByCarwashHandler handles GET /api/payments/carwash/{id}
func (tc *TransactionController) GetPaymentsByCarwashHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["id"]
	history, err := tc.TransactionService.GetTransactionHistory(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve carwash payment history")
		return
	}

	utils.JSON(w, http.StatusOK, history)
}
