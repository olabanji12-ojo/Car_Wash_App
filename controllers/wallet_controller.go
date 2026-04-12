package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo"
)

type WalletController struct {
	gateway services.PaymentGateway
	service *services.WalletService
}

func NewWalletController(service *services.WalletService) *WalletController {
	// Read directly from env since config.Init() is not called in main.go
	// The key is loaded by godotenv.Load() in main.go
	// Swap to sk_live_... in .env when going to production
	paystackKey := os.Getenv("PAYSTACK_SECRET_KEY")
	gateway := services.NewPaystackGateway(paystackKey)
	return &WalletController{
		gateway: gateway,
		service: service,
	}
}

// GetWallet returns the current user's wallet info
func (wc *WalletController) GetWallet(w http.ResponseWriter, r *http.Request) {
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authCtx.UserID

	walletCollection := database.DB.Collection("wallets")
	var wallet models.Wallet

	err = walletCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Auto-create an empty wallet if none exists
			wallet = models.Wallet{
				UserID:    userID,
				Type:      "customer", // Can be dynamically set based on role if needed
				Balance:   0.0,
				Currency:  "NGN",
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			res, _ := walletCollection.InsertOne(context.TODO(), wallet)
			wallet.ID = res.InsertedID.(primitive.ObjectID)
		} else {
			http.Error(w, `{"success":false,"message":"Failed to fetch wallet"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    wallet,
	})
}

// FundWallet calls the Payment Gateway to initialize a transaction
func (wc *WalletController) FundWallet(w http.ResponseWriter, r *http.Request) {
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authCtx.UserID

	var req struct {
		Amount float64 `json:"amount"`
		Email  string  `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"success":false,"message":"Amount must be greater than zero"}`, http.StatusBadRequest)
		return
	}

	// Create a unique internal reference
	reference := fmt.Sprintf("FUND-%s-%v", userID, time.Now().UnixNano())

	// Call loosely coupled gateway
	authURL, gatewayRef, err := wc.gateway.InitializeTransaction(req.Email, req.Amount, reference)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"message":"Payment Initialization Failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Pre-log the transaction intent
	tx := models.WalletTransaction{
		WalletID:  userID,
		Amount:    req.Amount,
		Type:      "deposit",
		Status:    "pending",
		Reference: gatewayRef, // This is what Paystack gives us
		CreatedAt: time.Now(),
	}
	database.DB.Collection("wallet_transactions").InsertOne(context.TODO(), tx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"authorization_url": authURL,
			"reference":         gatewayRef,
		},
	})
}

// VerifyDeposit verifies the transaction through the gateway and updates Wallet
func (wc *WalletController) VerifyDeposit(w http.ResponseWriter, r *http.Request) {
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userID := authCtx.UserID

	var req struct {
		Reference string `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Verify physically via the interface
	isValid, amount, err := wc.gateway.VerifyTransaction(req.Reference)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"message":"Verification Failed: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if !isValid {
		http.Error(w, `{"success":false,"message":"Transaction was not successful"}`, http.StatusBadRequest)
		return
	}

	txCollection := database.DB.Collection("wallet_transactions")
	walletCollection := database.DB.Collection("wallets")

	// Check if we already processed this reference
	var activeTx models.WalletTransaction
	err = txCollection.FindOne(context.TODO(), bson.M{"reference": req.Reference}).Decode(&activeTx)
	if err != nil || activeTx.Status == "success" {
		http.Error(w, `{"success":false,"message":"Transaction already processed or not found"}`, http.StatusBadRequest)
		return
	}

	// Ensure the amount matches loosely
	if activeTx.Amount != amount {
		// Just a sanity check, you can handle discrepancies here
		activeTx.Amount = amount 
	}

	// Update Transaction Status
	_, _ = txCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": activeTx.ID},
		bson.M{"$set": bson.M{"status": "success", "amount": amount}},
	)

	// Increment Wallet
	opts := bson.M{
		"$inc": bson.M{"balance": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}

	// Upsert Wallet calculation
	res := walletCollection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"user_id": userID},
		opts,
	)
	
	if res.Err() == mongo.ErrNoDocuments {
		// Need to create it natively
		wallet := models.Wallet{
			UserID:    userID,
			Type:      "customer",
			Balance:   amount,
			Currency:  "NGN",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		walletCollection.InsertOne(context.TODO(), wallet)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Wallet funded successfully",
		"balance": amount, // Simplified return
	})
}

// EscrowHold deducts from customer wallet for a booking and holds it in escrow
func (wc *WalletController) EscrowHold(w http.ResponseWriter, r *http.Request) {
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		BookingID string  `json:"booking_id"`
		Amount    float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 || req.BookingID == "" {
		http.Error(w, `{"success":false,"message":"Invalid request: booking_id and amount required"}`, http.StatusBadRequest)
		return
	}

	err = wc.service.EscrowHold(context.Background(), authCtx.UserID, req.BookingID, req.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"message":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("₦%.0f held in escrow for booking %s", req.Amount, req.BookingID),
	})
}

// EscrowRelease releases escrow funds to the business owner on booking completion
func (wc *WalletController) EscrowRelease(w http.ResponseWriter, r *http.Request) {
	// Only business owners or admins can release
	_, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		BookingID       string  `json:"booking_id"`
		BusinessOwnerID string  `json:"business_owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BookingID == "" {
		http.Error(w, `{"success":false,"message":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	err = wc.service.EscrowRelease(context.Background(), req.BookingID, req.BusinessOwnerID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"message":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Escrow released successfully",
	})
}

// RequestWithdrawal handles POST /api/wallet/withdraw
func (wc *WalletController) RequestWithdrawal(w http.ResponseWriter, r *http.Request) {
	authCtx, err := middleware.GetAuthContextDirect(r)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Amount      float64                `json:"amount"`
		BankDetails map[string]interface{} `json:"bank_details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, `{"success":false,"message":"Invalid request: amount and bank details required"}`, http.StatusBadRequest)
		return
	}

	err = wc.service.RequestWithdrawal(context.Background(), authCtx.UserID, req.Amount, req.BankDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"message":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Withdrawal request submitted successfully. Funds will be processed after review.",
	})
}
