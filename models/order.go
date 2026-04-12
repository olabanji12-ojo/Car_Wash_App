package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	
)

type Order struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	BookingID     primitive.ObjectID   `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	LinkedOrderID *primitive.ObjectID  `bson:"linked_order_id,omitempty" json:"linked_order_id,omitempty"` // For multi-car bookings
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
	PlatformFee   float64              `bson:"platform_fee" json:"platform_fee"`   // The 5% share
	MerchantNet   float64              `bson:"merchant_net" json:"merchant_net"`   // What the boss gets
	PaymentStatus string               `bson:"payment_status" json:"payment_status"` // paid / unpaid

	// Home service fields (optional copy from booking)
	BookingType  string       `bson:"booking_type,omitempty" json:"booking_type,omitempty"`
	UserLocation *GeoLocation `bson:"user_location,omitempty" json:"user_location,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	CompletionPhotoURL string `bson:"completion_photo_url,omitempty" json:"completion_photo_url,omitempty"`
	CompletionNote     string `bson:"completion_note,omitempty" json:"completion_note,omitempty"`
}


func (o Order) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.BookingID, validation.Required),
		validation.Field(&o.UserID, validation.Required),
		validation.Field(&o.CarID, validation.Required),
		validation.Field(&o.CarwashID, validation.Required),
		validation.Field(&o.ServiceIDs, validation.Required),
		validation.Field(&o.Status, validation.Required, validation.In("active", "completed")),
		validation.Field(&o.PaymentStatus, validation.Required, validation.In("paid", "unpaid")),
	)
}


