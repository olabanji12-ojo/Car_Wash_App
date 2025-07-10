package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


type Service struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CarwashID   primitive.ObjectID `bson:"carwash_id" json:"carwash_id"`       // Linked to the carwash
	Name        string             `bson:"name" json:"name"`                   // e.g., "Full Wash"
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Price       float64            `bson:"price" json:"price"`                 // e.g., 3500.00
	Duration    int                `bson:"duration" json:"duration"`           // In minutes
	IsAddon     bool               `bson:"is_addon" json:"is_addon"`           // Optional add-on service?
	Active      bool               `bson:"active" json:"active"`               // Still being offered?
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}




