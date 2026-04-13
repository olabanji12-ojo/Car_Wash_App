package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/sirupsen/logrus"
)

type AdminController struct {
	WalletService *services.WalletService
}

func NewAdminController(walletService *services.WalletService) *AdminController {
	return &AdminController{WalletService: walletService}
}

// ApprovePayout handles the administrative approval of a withdrawal request
func (ac *AdminController) ApprovePayout(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if input.TransactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	err := ac.WalletService.ApproveWithdrawal(context.Background(), input.TransactionID)
	if err != nil {
		logrus.Errorf("❌ Admin Payout Approval failed for TX %s: %v", input.TransactionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payout approved and transfer initiated successfully",
	})
}
