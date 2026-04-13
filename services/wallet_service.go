package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WalletService struct {
	repo           *repositories.WalletRepository
	financeService *FinanceService
	gateway        PaymentGateway
}

func NewWalletService(repo *repositories.WalletRepository, financeService *FinanceService, gateway PaymentGateway) *WalletService {
	return &WalletService{
		repo:           repo,
		financeService: financeService,
		gateway:        gateway,
	}
}

// EscrowHold locks funds for a specific booking
func (ws *WalletService) EscrowHold(ctx context.Context, userID string, bookingID string, amount float64) error {
	// 1. Get wallet
	wallet, err := ws.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return errors.New("wallet not found. Please fund your wallet first")
	}

	if wallet.Balance < amount {
		return fmt.Errorf("insufficient balance. Required: ₦%.2f, Available: ₦%.2f", amount, wallet.Balance)
	}

	// 2. Deduct from wallet
	if err := ws.repo.UpdateBalance(ctx, userID, -amount); err != nil {
		return err
	}

	// 3. Record transaction
	tx := &models.WalletTransaction{
		WalletID:  userID,
		Amount:    amount,
		Type:      "escrow_hold",
		Status:    "escrow",
		Reference: fmt.Sprintf("ESCROW-%s", bookingID),
		Metadata:  map[string]interface{}{"booking_id": bookingID},
		CreatedAt: time.Now(),
	}
	
	return ws.repo.CreateTransaction(ctx, tx)
}

// EscrowRelease releases funds to the business owner
func (ws *WalletService) EscrowRelease(ctx context.Context, bookingID string, carwashOwnerID string) error {
	// 1. Find the escrow hold
	holdTx, err := ws.repo.FindEscrowTransaction(ctx, bookingID)
	if err != nil {
		return errors.New("active escrow transaction not found for this booking")
	}

	// 2. Calculate split
	platformFee, merchantNet := ws.financeService.CalculateSplit(holdTx.Amount)

	// 3. Credit carwash owner
	if err := ws.repo.UpdateBalance(ctx, carwashOwnerID, merchantNet); err != nil {
		return err
	}

	// 4. Mark original escrow as released
	if err := ws.repo.UpdateTransactionStatus(ctx, holdTx.ID, "released"); err != nil {
		return err
	}

	// 5. Record the release transaction for the carwash
	releaseTx := &models.WalletTransaction{
		WalletID:  carwashOwnerID,
		Amount:    merchantNet,
		Type:      "escrow_release",
		Status:    "success",
		Reference: fmt.Sprintf("RELEASE-%s", bookingID),
		Metadata: map[string]interface{}{
			"booking_id":   bookingID,
			"platform_fee": platformFee,
		},
		CreatedAt: time.Now(),
	}
	
	return ws.repo.CreateTransaction(ctx, releaseTx)
}

// RequestWithdrawal initiates a withdrawal process for a user
func (ws *WalletService) RequestWithdrawal(ctx context.Context, userID string, amount float64, bankDetails map[string]interface{}) error {
	if amount <= 0 {
		return fmt.Errorf("invalid amount")
	}

	wallet, err := ws.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if wallet.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Create a transaction record
	tx := &models.WalletTransaction{
		WalletID:  userID,
		Amount:    amount,
		Type:      "payout",
		Status:    "pending",
		Reference: fmt.Sprintf("WD-%s-%d", userID, time.Now().Unix()),
		Metadata:  bankDetails,
		CreatedAt: time.Now(),
	}

	if err := ws.repo.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	// Deduct balance
	if err := ws.repo.UpdateBalance(ctx, userID, -amount); err != nil {
		return err
	}

	return nil
}

// EscrowRefund returns funds to the customer after a cancellation
func (ws *WalletService) EscrowRefund(ctx context.Context, bookingID string) error {
	// 1. Find the escrow hold
	holdTx, err := ws.repo.FindEscrowTransaction(ctx, bookingID)
	if err != nil {
		return errors.New("active escrow transaction not found for this booking")
	}

	// 2. Calculate refund (minus cancellation fee)
	refundAmount, fee := ws.financeService.CalculateRefundAmount(holdTx.Amount)

	// 3. Credit user balance
	if err := ws.repo.UpdateBalance(ctx, holdTx.WalletID, refundAmount); err != nil {
		return err
	}

	// 4. Mark original escrow as refunded
	if err := ws.repo.UpdateTransactionStatus(ctx, holdTx.ID, "refunded"); err != nil {
		return err
	}

	// 5. Record the refund transaction
	refundTx := &models.WalletTransaction{
		WalletID:  holdTx.WalletID,
		Amount:    refundAmount,
		Type:      "refund",
		Status:    "success",
		Reference: fmt.Sprintf("REFUND-%s", bookingID),
		Metadata: map[string]interface{}{
			"booking_id":       bookingID,
			"cancellation_fee": fee,
			"original_amount":  holdTx.Amount,
		},
		CreatedAt: time.Now(),
	}

	return ws.repo.CreateTransaction(ctx, refundTx)
}

// ApproveWithdrawal processes a Payout request and triggers the actual transfer
func (ws *WalletService) ApproveWithdrawal(ctx context.Context, txID string) error {
	objID, err := primitive.ObjectIDFromHex(txID)
	if err != nil {
		return errors.New("invalid transaction ID")
	}

	// 1. Find the transaction
	tx, err := ws.repo.FindTransactionByID(ctx, objID)
	if err != nil {
		return errors.New("transaction not found")
	}

	if tx.Type != "payout" || tx.Status != "pending" {
		return errors.New("only pending payout transactions can be approved")
	}

	// 2. Extract bank details from metadata
	bankName, _ := tx.Metadata["bank_name"].(string)
	accNumber, _ := tx.Metadata["account_number"].(string)
	accName, _ := tx.Metadata["account_name"].(string)

	if bankName == "" || accNumber == "" {
		return errors.New("incomplete bank details in transaction metadata")
	}

	// For MVP: We need a mapping from Bank Name to Paystack Bank Code
	// In production, you would use Paystack's /bank API or a lookup table.
	// For now, we'll try to find or create the recipient.
	bankCode := "011" // Defaulting to GTBank for test purposes, or lookup
	
	// 3. Create Transfer Recipient
	recipientCode, err := ws.gateway.CreateTransferRecipient(accName, accNumber, bankCode)
	if err != nil {
		return fmt.Errorf("failed to create paystack recipient: %v", err)
	}

	// 4. Initiate Transfer
	transferCode, err := ws.gateway.InitiateTransfer(tx.Amount, recipientCode, tx.Reference)
	if err != nil {
		return fmt.Errorf("paystack transfer initiation failed: %v", err)
	}

	// 5. Update transaction status
	updates := map[string]interface{}{
		"status": "success", 
		"metadata.transfer_code": transferCode,
		"metadata.processed_at": time.Now(),
	}
	
	// We'll use a direct repo update for metadata fields
	return ws.repo.UpdateTransactionFull(ctx, objID, updates)
}
