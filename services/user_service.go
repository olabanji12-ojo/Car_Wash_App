package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProfilePhotoFile represents an uploaded profile photo
type ProfilePhotoFile struct {
	File     multipart.File
	Filename string
	Size     int64
}

// UserService handles business logic for user operations
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user's profile using their ID
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	// 1. Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// 2. Fetch user from DB using repository
	user, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Sanitize sensitive fields
	user.Password = "" // Don't expose hashed password

	return user, nil
}

// UpdateUser updates basic profile info (NO FILE UPLOAD - for backward compatibility)
func (s *UserService) UpdateUser(userID string, input *models.User) (*models.User, error) {
	return s.UpdateUserWithPhoto(userID, input, nil)
}

// UpdateUserWithPhoto updates profile info including optional photo upload
func (s *UserService) UpdateUserWithPhoto(userID string, input *models.User, photoFile *ProfilePhotoFile) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Get current user data (we might need the old profile photo)
	currentUser, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Build update fields
	update := bson.M{}
	
	// Update basic fields
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Phone != "" {
		update["phone"] = input.Phone
	}
	if input.Email != "" {
		update["email"] = input.Email
	}

	// 4. Handle profile photo upload if provided
	if photoFile != nil {
		// Validate file
		if err := s.validateImageFile(photoFile); err != nil {
			return nil, err
		}

		// Generate unique filename
		filename := s.generateUniqueFilename(userID, photoFile.Filename)
		
		// Upload to Cloudinary
		uploadResult, err := UploadImage(photoFile.File, filename, "profile_photos")
		if err != nil {
			return nil, fmt.Errorf("failed to upload profile photo: %v", err)
		}

		// Add photo URL to update
		update["profile_photo"] = uploadResult.SecureURL

		// Delete old profile photo if it exists
		if currentUser.ProfilePhoto != "" {
			go func() {
				// Extract public_id from old URL and delete asynchronously
				// This prevents blocking the user update if deletion fails
				oldPublicID := s.extractPublicIDFromURL(currentUser.ProfilePhoto)
				if oldPublicID != "" {
					DeleteImage(oldPublicID)
				}
			}()
		}
	}

	// Always update timestamp
	update["updated_at"] = time.Now()

	// 5. Update in DB using repository
	err = s.userRepo.UpdateUserByID(objID, update)
	if err != nil {
		// If DB update fails but we uploaded a new image, we should clean it up
		if photoFile != nil && update["profile_photo"] != nil {
			// Try to delete the uploaded image
			filename := s.generateUniqueFilename(userID, photoFile.Filename)
			DeleteImage(filename) // Best effort cleanup
		}
		return nil, err
	}

	// 6. Get updated user
	updatedUser, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return nil, err
	}

	// 7. Hide password
	updatedUser.Password = ""

	return updatedUser, nil
}

// validateImageFile checks if the uploaded file is valid
func (s *UserService) validateImageFile(photoFile *ProfilePhotoFile) error {
	// Check file size (5MB limit)
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if photoFile.Size > maxSize {
		return errors.New("file size too large (max 5MB)")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(photoFile.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	
	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}
	
	if !isValidExt {
		return errors.New("invalid file type (allowed: jpg, jpeg, png, gif, webp)")
	}

	return nil
}

// generateUniqueFilename creates a unique filename for Cloudinary
func (s *UserService) generateUniqueFilename(userID, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("user_%s_%d%s", userID, timestamp, ext)
}

// extractPublicIDFromURL extracts Cloudinary public_id from URL for deletion
func (s *UserService) extractPublicIDFromURL(url string) string {
	// Cloudinary URLs typically look like:
	// https://res.cloudinary.com/your_cloud/image/upload/v1234567890/profile_photos/user_123_1234567890.jpg
	// We need to extract: profile_photos/user_123_1234567890
	
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}
	
	// Find the upload part
	uploadIndex := -1
	for i, part := range parts {
		if part == "upload" {
			uploadIndex = i
			break
		}
	}
	
	if uploadIndex == -1 || uploadIndex+3 >= len(parts) {
		return ""
	}
	
	// Skip version (v1234567890) and get folder/filename
	folder := parts[uploadIndex+2]
	filename := parts[uploadIndex+3]
	
	// Remove extension
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	
	return folder + "/" + filename
}

// DeleteProfilePhoto deletes a user's profile photo from Cloudinary and the database
func (s *UserService) DeleteProfilePhoto(userID string) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Get current user data
	currentUser, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Check if a profile photo exists
	if currentUser.ProfilePhoto == "" {
		// No photo to delete, return user as is
		currentUser.Password = "" // Sanitize password
		return currentUser, nil
	}

	// 4. Extract Public ID and delete from Cloudinary
	publicID := s.extractPublicIDFromURL(currentUser.ProfilePhoto)
	if publicID != "" {
		err := DeleteImage(publicID)
		if err != nil {
			// Log the error but don't block the user update
			fmt.Printf("Failed to delete image from Cloudinary (publicID: %s): %v\n", publicID, err)
		}
	}

	// 5. Build update to clear photo URL in DB
	update := bson.M{
		"profile_photo": "",
		"updated_at":    time.Now(),
	}

	// 6. Update in DB
	err = s.userRepo.UpdateUserByID(objID, update)
	if err != nil {
		return nil, err
	}

	// 7. Get the updated user to return
	updatedUser, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return nil, err
	}

	// 8. Sanitize password and return
	updatedUser.Password = ""
	return updatedUser, nil
}

// DeleteUser performs a soft delete on a user
func (s *UserService) DeleteUser(userID string) error {
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
	err = s.userRepo.UpdateUserByID(objID, update)
	if err != nil {
		return err
	}

	return nil
}

// GetUserRole retrieves the role of a user
func (s *UserService) GetUserRole(userID string) (string, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", errors.New("invalid user ID format")
	}

	// 2. Fetch user from DB
	user, err := s.userRepo.FindUserByID(objID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 3. Return the role
	return user.Role, nil
}

// GetLoyaltyPoints retrieves loyalty points for a car owner
func (s *UserService) GetLoyaltyPoints(userID string) (int, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	// 2. Fetch user
	user, err := s.userRepo.FindUserByID(objID)
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

// GetPublicProfile retrieves public profile information
func (s *UserService) GetPublicProfile(userID string) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Fetch user
	user, err := s.userRepo.FindUserByID(objID)
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

// UpdateUserCarwashID updates the carwash ID for a user
func (s *UserService) UpdateUserCarwashID(userID, carwashID primitive.ObjectID) error {
	return s.userRepo.UpdateUserCarwashID(userID, carwashID)
}

// UpdateUserCarwashIDByString is a helper function that accepts string IDs
func (s *UserService) UpdateUserCarwashIDByString(userIDStr, carwashIDStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	carwashID, err := primitive.ObjectIDFromHex(carwashIDStr)
	if err != nil {
		return errors.New("invalid carwash ID format")
	}

	return s.userRepo.UpdateUserCarwashID(userID, carwashID)
}

// AddAddress adds a new address to a user's profile
func (s *UserService) AddAddress(userID primitive.ObjectID, address models.UserAddress) error {
	// Validate the address
	if err := address.Validate(); err != nil {
		return err
	}

	// Ensure the address has an ID if not provided
	if address.ID == "" {
		address.ID = primitive.NewObjectID().Hex()
	}

	return s.userRepo.AddAddress(userID, address)
}

// UpdateAddress updates an existing address
func (s *UserService) UpdateAddress(userID primitive.ObjectID, addressID string, updates map[string]interface{}) error {
	// Validate updates
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	// If updating to default, ensure only one default exists
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		// Get current default address
		defaultAddr, err := s.userRepo.GetDefaultAddress(userID)
		if err == nil && defaultAddr != nil && defaultAddr.ID != addressID {
			// Update the old default to false
			if err := s.userRepo.UpdateAddress(userID, defaultAddr.ID, map[string]interface{}{
				"is_default": false,
			}); err != nil {
				return err
			}
		}
	}

	return s.userRepo.UpdateAddress(userID, addressID, updates)
}

// DeleteAddress removes an address from a user's profile
func (s *UserService) DeleteAddress(userID primitive.ObjectID, addressID string) error {
	// Check if the address exists and is not the last one
	addresses, err := s.userRepo.GetUserAddresses(userID)
	if err != nil {
		return err
	}

	if len(addresses) <= 1 {
		return errors.New("cannot delete the only address")
	}

	return s.userRepo.DeleteAddress(userID, addressID)
}

// GetUserAddresses retrieves all addresses for a user
func (s *UserService) GetUserAddresses(userID primitive.ObjectID) ([]models.UserAddress, error) {
	return s.userRepo.GetUserAddresses(userID)
}

// UpdateUserLocation updates the user's last known location
func (s *UserService) UpdateUserLocation(userID primitive.ObjectID, lat, lng float64) error {
	location := models.NewGeoPoint(lng, lat)
	return s.userRepo.UpdateUserLocation(userID, location)
}

// GetDefaultAddress returns the user's default address
func (s *UserService) GetDefaultAddress(userID primitive.ObjectID) (*models.UserAddress, error) {
	return s.userRepo.GetDefaultAddress(userID)
}

// SetDefaultAddress sets an address as the default
func (s *UserService) SetDefaultAddress(userID primitive.ObjectID, addressID string) error {
	// First, find the address to ensure it exists
	addresses, err := s.userRepo.GetUserAddresses(userID)
	if err != nil {
		return err
	}

	var found bool
	for _, addr := range addresses {
		if addr.ID == addressID {
			found = true
			break
		}
	}

	if !found {
		return errors.New("address not found")
	}

	// Update the address to be default
	return s.userRepo.UpdateAddress(userID, addressID, map[string]interface{}{
		"is_default": true,
	})
}