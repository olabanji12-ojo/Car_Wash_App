package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"

)

type Review struct {

	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`             // Who submitted the review
	CarwashID    primitive.ObjectID `bson:"carwash_id" json:"carwash_id"`       // Which carwash
	OrderID      *primitive.ObjectID `bson:"order_id,omitempty" json:"order_id,omitempty"` // Optional order link
	Rating       int                `bson:"rating" json:"rating"`               // Overall rating
	Accuracy     int                `bson:"accuracy" json:"accuracy"`           // Description match
	Cleanliness  int                `bson:"cleanliness" json:"cleanliness"`     // Was it clean?
	WorkerID     int                `bson:"worker_id,omitempty" json:"worker_id,omitempty"` // Optional
	Comment      string             `bson:"comment,omitempty" json:"comment,omitempty"`           // Freeform text
	Photos       []string           `bson:"photos,omitempty" json:"photos,omitempty"`             // Image URLs
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`

}


func (r Review) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required),
		validation.Field(&r.CarwashID, validation.Required),
		validation.Field(&r.Rating, validation.Required, validation.Min(1), validation.Max(5)),
		validation.Field(&r.Comment, validation.Length(0, 500)),
	)
}

