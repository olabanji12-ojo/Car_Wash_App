package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
)

type WalletService struct {
	repo           *repositories.WalletRepository
	financeService *FinanceService
}

func NewWalletService(repo *repositories.WalletRepository, financeService *FinanceService) *WalletService {
	return &WalletService{
		repo:           repo,
		financeService: financeService,
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
