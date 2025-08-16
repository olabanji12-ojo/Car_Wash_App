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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
)

// ProfilePhotoFile represents an uploaded profile photo
type ProfilePhotoFile struct {
	File     multipart.File
	Filename string
	Size     int64
}

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

// UpdateUser updates basic profile info (NO FILE UPLOAD - for backward compatibility)
func UpdateUser(userID string, input *models.User) (*models.User, error) {
	return UpdateUserWithPhoto(userID, input, nil)
}

// UpdateUserWithPhoto updates profile info including optional photo upload
func UpdateUserWithPhoto(userID string, input *models.User, photoFile *ProfilePhotoFile) (*models.User, error) {
	// 1. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// 2. Get current user data (we might need the old profile photo)
	currentUser, err := repositories.FindUserByID(objID)
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
		if err := validateImageFile(photoFile); err != nil {
			return nil, err
		}

		// Generate unique filename
		filename := generateUniqueFilename(userID, photoFile.Filename)
		
		// Upload to Cloudinary
		uploadResult, err := UploadImage(photoFile.File, filename, "profile_photos")
		if err != nil {
			return nil, fmt.Errorf("failed to upload profile photo: %v", err)
		}

		// Add photo URL to update
		update["profile_photo"] = uploadResult.SecureURL

		// TODO: Delete old profile photo if it exists
		// This is optional - you might want to keep old photos for a while
		if currentUser.ProfilePhoto != "" {
			go func() {
				// Extract public_id from old URL and delete asynchronously
				// This prevents blocking the user update if deletion fails
				oldPublicID := extractPublicIDFromURL(currentUser.ProfilePhoto)
				if oldPublicID != "" {
					DeleteImage(oldPublicID)
				}
			}()
		}
	}

	// Always update timestamp
	update["updated_at"] = time.Now()

	// 5. Update in DB using repository
	err = repositories.UpdateUserByID(objID, update)
	if err != nil {
		// If DB update fails but we uploaded a new image, we should clean it up
		if photoFile != nil && update["profile_photo"] != nil {
			// Try to delete the uploaded image
			filename := generateUniqueFilename(userID, photoFile.Filename)
			DeleteImage(filename) // Best effort cleanup
		}
		return nil, err
	}

	// 6. Get updated user
	updatedUser, err := repositories.FindUserByID(objID)
	if err != nil {
		return nil, err
	}

	// 7. Hide password
	updatedUser.Password = ""

	return updatedUser, nil
}

// validateImageFile checks if the uploaded file is valid
func validateImageFile(photoFile *ProfilePhotoFile) error {
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
func generateUniqueFilename(userID, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("user_%s_%d%s", userID, timestamp, ext)
}

// extractPublicIDFromURL extracts Cloudinary public_id from URL for deletion
func extractPublicIDFromURL(url string) string {
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