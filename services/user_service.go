package services

import (

	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"

)

// GetUserByID retrieves a user's profile using their ID
func GetUserByID(userID string) (*models.User, error) {
	// 1. Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// 2. Fetch user from DB using repository
	user, err := repositories.FindUserByID(objID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Sanitize sensitive fields
	user.Password = "" // Don't expose hashed password

	return user, nil
}


// Input struct for updates


// UpdateUser updates basic profile info
func UpdateUser(userID string, input *models.User) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Build update fields

	update := bson.M{}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Phone != "" {
		update["phone"] = input.Phone
	}
	if input.ProfilePhoto != "" {
		update["profile_photo"] = input.ProfilePhoto
	}
	update["updated_at"] = time.Now()

	// 3. Update in DB using repository
	err = repositories.UpdateUserByID(objID, update)
	if err != nil {
		return nil, err
	}

	// 4. Get updated user
	updatedUser, err := repositories.FindUserByID(objID)
	if err != nil {
		return nil, err
	}

	// 5. Hide password
	updatedUser.Password = ""

	return updatedUser, nil
}


//  DeleteUser  Function(performing a soft delete)

func DeleteUser(userID string) error {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// 2. Build soft delete update
	update := bson.M{
		"status":     "deleted",
		"updated_at": time.Now(),
	}

	// 3. Call repo to update
	err = repositories.UpdateUserByID(objID, update)
	if err != nil {
		return err
	}

	return nil
}


// GetUserRole Function

func GetUserRole(userID string) (string, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", errors.New("invalid user ID format")
	}

	// 2. Fetch user from DB
	user, err := repositories.FindUserByID(objID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 3. Return the role
	return user.Role, nil
}


// GetLoyaltyPoints Function

func GetLoyaltyPoints(userID string) (int, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	// 2. Fetch user
	user, err := repositories.FindUserByID(objID)
	if err != nil {
		return 0, errors.New("user not found")
	} 

	// 3. Confirm they are a car owner
	if user.Role != "car_owner" && user.AccountType != "car_owner" {
		return 0, errors.New("loyalty points not available for this user")
	}

	// 4. Return the loyalty points
	return user.LoyaltyPoints, nil
}


// 

func GetPublicProfile(userID string) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Fetch user
	user, err := repositories.FindUserByID(objID)
	if err != nil {
		return nil, err
	}

	// 3. Build public response
	public := &models.User{
		ID:           user.ID,
		Name:         user.Name,
		ProfilePhoto: user.ProfilePhoto,
		Role:         user.Role,
	}

	if user.AccountType == "car_wash" && user.Role == "worker" {
		public.JobRole = user.JobRole
	}

	return public, nil 
}
  
func UpdateUserCarwashID(userID, carwashID primitive.ObjectID) error {
	
	return repositories.UpdateUserCarwashID(userID, carwashID)
}  
 
