package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Wallet represents a digital wallet for a User or Business
type Wallet struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"` // Can be a customer ID or carwash ID
	Type      string             `bson:"type" json:"type"`       // "customer" or "carwash"
	Balance   float64            `bson:"balance" json:"balance"` // Always stored in minor units internally if possible, but here float for ease
	Currency  string             `bson:"currency" json:"currency"` // "NGN"
	Status    string             `bson:"status" json:"status"`     // "active", "locked"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// WalletTransaction tracks the money movement
type WalletTransaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WalletID  string             `bson:"wallet_id" json:"wallet_id"`
	Amount    float64            `bson:"amount" json:"amount"`
	Type      string             `bson:"type" json:"type"` // "deposit", "escrow_hold", "escrow_release", "payout", "refund"
	Status    string             `bson:"status" json:"status"` // "pending", "success", "failed"
	BookingID string             `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	Reference string             `bson:"reference" json:"reference"` // E.g., Paystack reference
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
