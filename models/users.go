package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/olabanji12-ojo/CarWashApp/utils" 


)

// A user account. All users (workers, business owners, car owners) will have a User document.
// If a business signs up, a corresponding business document is created.
// If a worker is added to a business, a user is created with the business’s CarWashID linked.


type User struct {


	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string              `bson:"name,omitempty" json:"name,omitempty"`
	CarWashID    *primitive.ObjectID `bson:"carwash_id,omitempty" json:"carwash_id,omitempty"`
	Email         string              `bson:"email,omitempty" json:"email,omitempty"`
	Password      string              `bson:"password,omitempty" json:"password,omitempty"`
	Phone         string              `bson:"phone,omitempty" json:"phone,omitempty"`
	Role          string              `bson:"role,omitempty" json:"role,omitempty"`                 // car_owner, business_owner, worker, business_admin
	Status        string              `bson:"status,omitempty" json:"status,omitempty"`             // active, pending, suspended
	AccountType   string              `bson:"account_type,omitempty" json:"account_type,omitempty"` //   car_owner, car_wash
	Verified      bool                `bson:"verified,omitempty" json:"verified,omitempty"`
	ProfilePhoto  string              `bson:"profile_photo,omitempty" json:"profile_photo,omitempty"`
	LoyaltyPoints int                 `bson:"loyalty_points,omitempty" json:"loyalty_points,omitempty"`
	LastSeen      *time.Time          `bson:"last_seen,omitempty" json:"last_seen,omitempty"`
	JobRole       string              `bson:"job_role,omitempty" json:"job_role,omitempty"`
	WorkerStatus  string              `bson:"worker_status,omitempty" json:"worker_status,omitempty"` // active, inactive, on break etc
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
	ActiveOrders  []primitive.ObjectID `bson:"active_orders,omitempty" json:"active_orders,omitempty"` 
	
    
}


func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 100)),
		validation.Field(&u.Role, validation.Required, validation.In(
			utils.ROLE_CAR_OWNER,
			utils.ROLE_WORKER,  
			utils.ROLE_BUSINESS, 
		)),

		validation.Field(&u.AccountType, validation.When(
			u.Role != utils.ROLE_WORKER, // Only validate if not a worker
			validation.In(
				utils.ACCOUNT_TYPE_CAR_WASH,
				utils.ACCOUNT_TYPE_CAR_OWNER,
			),
		)),   
	
	)
}


































// type UserUpdateInput struct {
// 	Name         string `json:"name,omitempty"`
// 	Phone        string `json:"phone,omitempty"`
// 	ProfilePhoto string `json:"profile_photo,omitempty"`
// }



// type PublicUserProfile struct {
// 	ID           primitive.ObjectID `json:"id"`
// 	Name         string             `json:"name"`
// 	ProfilePhoto string             `json:"profile_photo,omitempty"`
// 	Role         string             `json:"role"`
// 	JobRole      string             `json:"job_role,omitempty"`
// }

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
