package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	BookingID     primitive.ObjectID   `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	UserID        primitive.ObjectID   `bson:"user_id" json:"user_id"`
	CarID         primitive.ObjectID   `bson:"car_id" json:"car_id"`
	CarwashID     primitive.ObjectID   `bson:"carwash_id" json:"carwash_id"`
	ServiceIDs    []primitive.ObjectID `bson:"service_ids" json:"service_ids"`
	WorkerID      *primitive.ObjectID  `bson:"worker_id,omitempty" json:"worker_id,omitempty"`
	StartTime     time.Time            `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime       time.Time            `bson:"end_time,omitempty" json:"end_time,omitempty"`
	QueueNumber   int                  `bson:"queue_number" json:"queue_number"`
	Status        string               `bson:"status" json:"status"` // active, completed
	TotalAmount   float64              `bson:"total_amount" json:"total_amount"`
	PaymentStatus string               `bson:"payment_status" json:"payment_status"` // paid / unpaid

	//  Home service fields (optional copy from booking)
	BookingType  string       `bson:"booking_type,omitempty" json:"booking_type,omitempty"`
	UserLocation *GeoLocation `bson:"user_location,omitempty" json:"user_location,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}



