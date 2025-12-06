package repositories

import (
	"context"
	// "errors"
	"time"

	"fmt"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the MongoDB users collection
func (ur *UserRepository) CreateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ur.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		logrus.Error("Error inserting user: ", err)
		return err
	}
	logrus.Info("User inserted successfully")
	return nil

}

// FindUserByEmail searches for a user by their email
func (ur *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	logrus.Info("Searching for user by email in MongoDB")
	err := database.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		logrus.Warn("User not found with email: ", email)
		return nil, err
	}
	logrus.Info("Inserted user Successfully")
	return &user, nil
}

// FindUserByID fetches a user by MongoDB ObjectID
func (ur *UserRepository) FindUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := ur.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// New: Update user profile fields
func (ur *UserRepository) UpdateUserByID(userID primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ur.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},  // filter
		bson.M{"$set": update}, // update operation
	)

	return err
}

func (ur *UserRepository) UpdateUserCarwashID(userID, carwashID primitive.ObjectID) error {
	collection := ur.db.Collection("users")

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"carwash_id": carwashID}}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// AddAddress adds a new address to a user's profile
func (ur *UserRepository) AddAddress(userID primitive.ObjectID, address models.UserAddress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set created at time
	address.CreatedAt = time.Now()

	// Get the user to check for existing default address
	user, err := ur.FindUserByID(userID)
	if err != nil {
		return err
	}

	// If no default address exists, make this one default
	if user.DefaultAddressID == nil || *user.DefaultAddressID == "" {
		address.IsDefault = true
	}

	// Add the address to the user's addresses array
	update := bson.M{
		"$push": bson.M{"addresses": address},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	// If this is the first address or it's marked as default, update default address
	if address.IsDefault {
		update["$set"].(bson.M)["default_address_id"] = address.ID
	}

	_, err = ur.db.Collection("users").UpdateByID(ctx, userID, update)
	return err
}

// UpdateAddress updates an existing address
func (ur *UserRepository) UpdateAddress(userID primitive.ObjectID, addressID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build the update query
	setValues := bson.M{"updated_at": time.Now()}
	for key, value := range updates {
		setValues["addresses.$."+key] = value
	}

	// Find the user and update the specific address
	filter := bson.M{
		"_id":           userID,
		"addresses._id": addressID,
	}

	update := bson.M{
		"$set": setValues,
	}

	result, err := ur.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("address not found")
	}

	// If this address is being set as default, update the default_address_id
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		_, err = ur.db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$set": bson.M{"default_address_id": addressID}},
		)
	}

	return err
}

// DeleteAddress removes an address from a user's profile
func (ur *UserRepository) DeleteAddress(userID primitive.ObjectID, addressID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a session for transaction
	session, err := ur.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Use transaction to ensure data consistency
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Get the user to check if this is the default address
		var user models.User
		err := ur.db.Collection("users").FindOne(sessCtx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			return nil, err
		}

		// Remove the address
		update := bson.M{
			"$pull": bson.M{"addresses": bson.M{"_id": addressID}},
			"$set":  bson.M{"updated_at": time.Now()},
		}

		_, err = ur.db.Collection("users").UpdateOne(sessCtx, bson.M{"_id": userID}, update)
		if err != nil {
			return nil, err
		}

		// If this was the default address, update the default
		if user.DefaultAddressID != nil && *user.DefaultAddressID == addressID {
			// Get the updated user to check remaining addresses
			var updatedUser models.User
			err := ur.db.Collection("users").FindOne(sessCtx, bson.M{"_id": userID}).Decode(&updatedUser)
			if err != nil {
				return nil, err
			}

			if len(updatedUser.Addresses) > 1 { // More than one because we haven't removed the address yet
				// Set the first address that's not being deleted as default
				var newDefault string
				for _, addr := range updatedUser.Addresses {
					if addr.ID != addressID {
						newDefault = addr.ID
						break
					}
				}

				if newDefault != "" {
					_, err = ur.db.Collection("users").UpdateOne(
						sessCtx,
						bson.M{"_id": userID},
						bson.M{"$set": bson.M{"default_address_id": newDefault}},
					)
				}
			} else {
				// No addresses left, clear the default
				_, err = ur.db.Collection("users").UpdateOne(
					sessCtx,
					bson.M{"_id": userID},
					bson.M{"$unset": bson.M{"default_address_id": ""}},
				)
			}
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// GetUserAddresses retrieves all addresses for a user
func (ur *UserRepository) GetUserAddresses(userID primitive.ObjectID) ([]models.UserAddress, error) {
	user, err := ur.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user.Addresses, nil
}

// UpdateUserLocation updates the user's last known location
func (ur *UserRepository) UpdateUserLocation(userID primitive.ObjectID, location models.GeoPoint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ur.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"last_location": location,
				"updated_at":    time.Now(),
			},
		},
	)
	return err
}

// GetDefaultAddress returns the user's default address
func (ur *UserRepository) GetDefaultAddress(userID primitive.ObjectID) (*models.UserAddress, error) {
	user, err := ur.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	if user.DefaultAddressID == nil || *user.DefaultAddressID == "" {
		return nil, fmt.Errorf("no default address found")
	}

	for _, addr := range user.Addresses {
		if addr.ID == *user.DefaultAddressID {
			return &addr, nil
		}
	}

	return nil, fmt.Errorf("default address not found in addresses array")
}

// GetUsersByIDs fetches multiple users by their IDs
func (ur *UserRepository) GetUsersByIDs(userIDs []primitive.ObjectID) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$in": userIDs}}
	cursor, err := ur.db.Collection("users").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
