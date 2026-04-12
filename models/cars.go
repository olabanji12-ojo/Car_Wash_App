package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Car struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID       primitive.ObjectID `bson:"owner_id" json:"owner_id"` // References the User
	Model         string             `bson:"model" json:"model"`       // e.g., Toyota Camry
	Size          string             `bson:"size" json:"size"`         // small, medium, large
	Plate         string             `bson:"plate,omitempty" json:"plate,omitempty"`
	Color         string             `bson:"color,omitempty" json:"color,omitempty"`
	Profile_photo string             `bson:"profile_photo,omitempty" json:"profile_photo,omitempty"`
	IsDefault     bool               `bson:"is_default" json:"is_default"` // Used for quicker booking
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	Note          string             `bson:"note,omitempty" json:"note,omitempty"`
}

// adding the validation package
func (c Car) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Model, validation.Required),
		validation.Field(&c.Size, validation.Required, validation.In("small", "medium", "large")),
	)
}


