package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	CarwashID       primitive.ObjectID `bson:"carwash_id" json:"carwash_id"`
	OrderID         primitive.ObjectID `bson:"order_id" json:"order_id"`
	Amount          float64            `bson:"amount" json:"amount"`
	Method          string             `bson:"method" json:"method"` // card, cash, wallet, transfer
	Status          string             `bson:"status" json:"status"` // paid, failed, pending, refunded
	TransactionRef  string             `bson:"transaction_ref,omitempty" json:"transaction_ref,omitempty"`
	PaidAt          time.Time          `bson:"paid_at" json:"paid_at"` // When payment was actually made
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}


