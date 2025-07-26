package models

import (
	"errors"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a user notification
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title     string            `bson:"title" json:"title"`
	Message   string            `bson:"message" json:"message"`
	Type      string            `bson:"type" json:"type"` // "booking", "order", "payment", "worker"
	IsRead    bool              `bson:"is_read" json:"is_read"`
	EmailSent bool              `bson:"email_sent" json:"email_sent"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
}

// NotificationTypes - constants for notification types

const (

	NotificationTypeBooking = "booking"
	NotificationTypeOrder   = "order"
	NotificationTypePayment = "payment"
	NotificationTypeWorker  = "worker"
	NotificationTypeGeneral = "general"
	
)

// Validate validates the notification data
func (n *Notification) Validate() error {
	if n.UserID.IsZero() {
		return errors.New("user ID is required")
	}
	if n.Title == "" {
		return errors.New("title is required")
	}
	if n.Message == "" {
		return errors.New("message is required")
	}
	if n.Type == "" {
		return errors.New("type is required")
	}
	return nil
}
