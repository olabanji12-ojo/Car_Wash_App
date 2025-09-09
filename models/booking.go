package models


import (
	"time"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)



type Booking struct {
    	
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID   `bson:"user_id" json:"user_id"`
	CarID        primitive.ObjectID   `bson:"car_id" json:"car_id"`
	CarwashID    primitive.ObjectID   `bson:"carwash_id" json:"carwash_id"`
	
	BookingTime  time.Time            `bson:"booking_time" json:"booking_time"`
	BookingType  string               `bson:"booking_type" json:"booking_type"` // slot_booking / home_service
	UserLocation *GeoLocation         `bson:"user_location,omitempty" json:"user_location,omitempty"` // Only for home service
	AddressNote  string               `bson:"address_note,omitempty" json:"address_note,omitempty"`   // Optional directions 
	Status       string               `bson:"status" json:"status"`               // pending, confirmed, etc
	Notes        string               `bson:"notes,omitempty" json:"notes,omitempty"`
	QueueNumber  int                  `bson:"queue_number" json:"queue_number"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
    
    
}


func (b Booking) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.UserID, validation.Required),
		validation.Field(&b.CarID, validation.Required),
		validation.Field(&b.CarwashID, validation.Required),
		validation.Field(&b.BookingTime, validation.Required),
		validation.Field(&b.BookingType, validation.Required, validation.In("slot_booking", "home_service")),
		// validation.Field(&b.Status, validation.Required, validation.In("pending", "confirmed", "completed", "cancelled")),

	) 

	// Conditional check for home service
	// add them to constants and consider using enums if you have the time
	// how to handle user errors and server errors
	
	if b.BookingType == "home_service" && b.UserLocation == nil {
		return errors.New("user location is required for home service bookings")
	}

		return err
}


func (b *Booking) SetDefaults() {
    b.CreatedAt = time.Now()
    b.UpdatedAt = time.Now()
    b.ID = primitive.NewObjectID()
    
    // Set default status if empty
    if b.Status == "" {
        b.Status = "pending"
    }
}
