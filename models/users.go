package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	Password     string             `bson:"password" json:"password"`
	Phone        string             `bson:"phone" json:"phone"`
	Role         string             `bson:"role" json:"role"`     // car_owner, business, worker
	Status       string             `bson:"status" json:"status"` // active, pending, suspended
	Verified     bool               `bson:"verified" json:"verified"`
	ProfilePhoto string             `bson:"profile_photo,omitempty" json:"profile_photo,omitempty"`

	OwnerData  *OwnerProfile  `bson:"owner_data,omitempty" json:"owner_data,omitempty"`
	WorkerData *WorkerProfile `bson:"worker_data,omitempty" json:"worker_data,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Role-specific data for car owners
type OwnerProfile struct {
	LoyaltyPoints int `bson:"loyalty_points,omitempty" json:"loyalty_points,omitempty"`
}

// Role-specific data for workers
type WorkerProfile struct {
	BusinessID string `bson:"business_id,omitempty" json:"business_id,omitempty"`
	JobRole    string `bson:"job_role,omitempty" json:"job_role,omitempty"`

	// Single status field that covers everything
	WorkerStatus string    `bson:"status,omitempty" json:"status,omitempty"` // "online", "offline", "busy", "on_break"
	LastSeen     time.Time `bson:"last_seen,omitempty" json:"last_seen,omitempty"`

	// Order tracking
	ActiveOrders []primitive.ObjectID `bson:"active_orders,omitempty" json:"active_orders,omitempty"`
	// Location     *GeoLocation         `bson:"location,omitempty" json:"location,omitempty"` // instances where worker is not at the carwash
}

// this is for a user account, every user that signs up will have a user account document
// a business can also sign up from the app, if a business signs up, a corresponding business document should be created for him
// if a worker is added to a business, a user document is created for that worker and the business id is updated to the business id of the admin that added him thereby mapping the worker to the existing business record
type UserUpdate struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string              `bson:"name,omitempty" json:"name,omitempty"`
	BusinessID    *primitive.ObjectID `bson:"business_id,omitempty" json:"business_id,omitempty"`
	Email         string              `bson:"email,omitempty" json:"email,omitempty"`
	Password      string              `bson:"password,omitempty" json:"password,omitempty"`
	Phone         string              `bson:"phone,omitempty" json:"phone,omitempty"`
	Role          string              `bson:"role,omitempty" json:"role,omitempty"`                 // car_owner, business_owner, worker, business_admin
	Status        string              `bson:"status,omitempty" json:"status,omitempty"`             // active, pending, suspended
	AccountType   string              `bson:"account_type,omitempty" json:"account_type,omitempty"` // worker, car_owner, car_wash
	Verified      bool                `bson:"verified,omitempty" json:"verified,omitempty"`
	ProfilePhoto  string              `bson:"profile_photo,omitempty" json:"profile_photo,omitempty"`
	LoyaltyPoints int                 `bson:"loyalty_points,omitempty" json:"loyalty_points,omitempty"`
	LastSeen      *time.Time          `bson:"last_seen,omitempty" json:"last_seen,omitempty"`
	JobRole       string              `bson:"job_role,omitempty" json:"job_role,omitempty"`
	WorkerStatus  string              `bson:"worker_status,omitempty" json:"worker_status,omitempty"` // active, inactive, on break etc
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
}

type UserUpdateInput struct {
	Name         string `json:"name,omitempty"`
	Phone        string `json:"phone,omitempty"`
	ProfilePhoto string `json:"profile_photo,omitempty"`
}

type PublicUserProfile struct {
	ID           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	ProfilePhoto string             `json:"profile_photo,omitempty"`
	Role         string             `json:"role"`
	JobRole      string             `json:"job_role,omitempty"`
}

// CreateWorkerInput Struct

// type CreateWorkerInput struct {
// 	Name     string `json:"name"`
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// 	Phone    string `json:"phone,omitempty"`
// 	JobRole  string `json:"job_role"`
// }

// func (input CreateWorkerInput) Validate() error {
// 	return validation.ValidateStruct(&input,
// 		validation.Field(&input.Name, validation.Required, validation.Length(2, 50)),
// 		validation.Field(&input.Email, validation.Required, is.Email),
// 		validation.Field(&input.Password, validation.Required, validation.Length(6, 100)),
// 		validation.Field(&input.JobRole, validation.Required),
// 	)
// }

func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 100)),
		validation.Field(&u.Role, validation.Required, validation.In("car_owner", "worker", "business")),
	)
}
