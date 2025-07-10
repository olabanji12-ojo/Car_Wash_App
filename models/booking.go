package models


import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)



type Booking struct {
    
	
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID   `bson:"user_id" json:"user_id"`
	CarID        primitive.ObjectID   `bson:"car_id" json:"car_id"`
	CarwashID    primitive.ObjectID   `bson:"carwash_id" json:"carwash_id"`
	ServiceIDs   []primitive.ObjectID `bson:"service_ids" json:"service_ids"`
	BookingTime  time.Time            `bson:"booking_time" json:"booking_time"`
	BookingType  string               `bson:"booking_type" json:"booking_type"` // walk_in / home_service
	UserLocation *GeoLocation         `bson:"user_location,omitempty" json:"user_location,omitempty"` // Only for home service
	AddressNote  string               `bson:"address_note,omitempty" json:"address_note,omitempty"`   // Optional directions
	Status       string               `bson:"status" json:"status"`               // pending, approved, etc
	Notes        string               `bson:"notes,omitempty" json:"notes,omitempty"`
	QueueNumber  int                  `bson:"queue_number" json:"queue_number"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
    
    
}



