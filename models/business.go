package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Business struct {
		ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		Name         string             `bson:"name" json:"name"`
		Email        string             `bson:"email" json:"email"`
		Phone        string             `bson:"phone" json:"phone"`
		Address      string             `bson:"address" json:"address"`
		Logo         string             `bson:"logo" json:"logo"`
		OpeningHours time.Time          `bson:"opening_hours" json:"opening_hours"`
		ClosingHours time.Time          `bson:"closing_hours" json:"closing_hours"`
		CreatedAt    time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
		State        string             `bson:"state,omitempty" json:"state,omitempty"`
		Country      string             `bson:"country,omitempty" json:"country,omitempty"`
		LGA          string             `bson:"lga,omitempty" json:"lga,omitempty"`
		UpdatedAt    *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	}
)

func (b *Business) SetDefaults() {
	b.CreatedAt = time.Now()
	b.ID = primitive.NewObjectID()
}
