package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"

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


func (p Payment) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.UserID, validation.Required),
		validation.Field(&p.CarwashID, validation.Required),
		validation.Field(&p.OrderID, validation.Required),
		validation.Field(&p.Amount, validation.Required),
		validation.Field(&p.Method, validation.Required, validation.In("cash", "card", "wallet", "transfer")),
		validation.Field(&p.Status, validation.Required, validation.In("paid", "failed", "pending", "refunded")),
	)
}

