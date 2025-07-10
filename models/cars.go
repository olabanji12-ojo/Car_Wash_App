package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Car struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID    primitive.ObjectID `bson:"owner_id" json:"owner_id"`                   // References the User
	Model      string             `bson:"model" json:"model"`                         // e.g., Toyota Camry
	Plate      string             `bson:"plate" json:"plate"`                         // e.g., ABC-1234
	Color      string             `bson:"color,omitempty" json:"color,omitempty"`     // Optional
	IsDefault  bool               `bson:"is_default" json:"is_default"`               // Used for quicker booking
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}



