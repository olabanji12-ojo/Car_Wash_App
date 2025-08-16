package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CarPhotoFile represents an uploaded car photo
type CarPhotoFile struct {
	File     multipart.File
	Filename string
	Size     int64
}

// CreateCar - Original function for backward compatibility (no photo)
func CreateCar(userID string, input models.Car) (*models.Car, error) {
	return CreateCarWithPhoto(userID, input, nil)
}


// CreateCarWithPhoto - NEW function that handles photo upload during creation
func CreateCarWithPhoto(userID string, input models.Car, photoFile *CarPhotoFile) (*models.Car, error) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout for file upload
	defer cancel()

	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// If this is set as default, unset others first
	if input.IsDefault {
		if err := repositories.UnsetDefaultCarsForUser(ownerID); err != nil {
			return nil, err
		}
	}


	// Handle photo upload if provided
	var profilePhotoURL string
	if photoFile != nil {
		// Validate file
		if err := validateCarImageFile(photoFile); err != nil {
			return nil, err
		}

		// Generate unique filename for car photo
		filename := generateCarPhotoFilename(userID, input.Plate, photoFile.Filename)
		
		// Upload to Cloudinary
		uploadResult, err := UploadImage(photoFile.File, filename, "car_photos")
		if err != nil {
			return nil, fmt.Errorf("failed to upload car photo: %v", err)
		}

		profilePhotoURL = uploadResult.SecureURL
	}


	// Create new car with photo URL
	newCar := models.Car{
		ID:            primitive.NewObjectID(),
		OwnerID:       ownerID,
		Model:         input.Model,
		Plate:         input.Plate,
		Color:         input.Color,
		Profile_photo: profilePhotoURL, // Set the uploaded photo URL
		IsDefault:     input.IsDefault,
		CreatedAt:     time.Now(),   
		UpdatedAt:     time.Now(),  
	}   

	// Save to database
	if err := repositories.CreateCar(&newCar); err != nil {
		// If DB save fails but we uploaded an image, clean it up
		if profilePhotoURL != "" {
			go func() {
				publicID := extractCarPhotoPublicIDFromURL(profilePhotoURL)
				if publicID != "" {
					DeleteImage(publicID)
				}
			}()
		}
		logrus.Error("Failed to create car: ", err)
		return nil, err
	}

	return &newCar, nil
}

// GetCarsByUserID - No changes needed, returns cars with photo URLs
func GetCarsByUserID(userID string) ([]models.Car, error) {
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetCarsByUserID(ownerID)
}

// UpdateCar - Original function for backward compatibility (no photo)
func UpdateCar(userID, carID string, updates map[string]interface{}) error {
	return UpdateCarWithPhoto(userID, carID, updates, nil)
}

// UpdateCarWithPhoto - NEW function that handles photo upload during updates
func UpdateCarWithPhoto(userID, carID string, updates map[string]interface{}, photoFile *CarPhotoFile) error {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	// Get current car data (we might need the old photo)
	currentCar, err := repositories.GetCarByID(carObjID)
	if err != nil {
		return errors.New("car not found")
	}

	// Handle photo upload if provided
	if photoFile != nil {
		// Validate file
		if err := validateCarImageFile(photoFile); err != nil {
			return err
		}

		// Generate unique filename for car photo
		filename := generateCarPhotoFilename(userID, currentCar.Plate, photoFile.Filename)
		
		// Upload to Cloudinary
		uploadResult, err := UploadImage(photoFile.File, filename, "car_photos")
		if err != nil {
			return fmt.Errorf("failed to upload car photo: %v", err)
		}

		// Add photo URL to updates
		updates["profile_photo"] = uploadResult.SecureURL

		// Clean up old photo if it exists
		if currentCar.Profile_photo != "" {
			go func() {
				// Delete old photo asynchronously
				oldPublicID := extractCarPhotoPublicIDFromURL(currentCar.Profile_photo)
				if oldPublicID != "" {
					DeleteImage(oldPublicID)
				}
			}()
		}
	}

	// Add updatedAt
	updates["updated_at"] = time.Now()

	// Update in database
	err = repositories.UpdateCar(carObjID, bson.M(updates))
	if err != nil {
		// If DB update fails but we uploaded a new image, clean it up
		if photoFile != nil && updates["profile_photo"] != nil {
			go func() {
				filename := generateCarPhotoFilename(userID, currentCar.Plate, photoFile.Filename)
				DeleteImage(filename) // Best effort cleanup
			}()
		}
		return err
	}

	return nil
}

// DeleteCar - Enhanced to clean up photos when deleting
func DeleteCar(userID, carID string) error {
	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	// Get car data before deletion to clean up photo
	car, err := repositories.GetCarByID(carObjID)
	if err == nil && car.Profile_photo != "" {
		// Delete photo from Cloudinary asynchronously after DB deletion
		go func() {
			publicID := extractCarPhotoPublicIDFromURL(car.Profile_photo)
			if publicID != "" {
				DeleteImage(publicID)
			}
		}()
	}

	// Delete from database
	return repositories.DeleteCar(carObjID)
}

// SetDefaultCar - No changes needed
func SetDefaultCar(userID, carID string) error {
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	carObjID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return errors.New("invalid car ID")
	}

	return repositories.SetDefaultCar(ownerID, carObjID)
}

// GetCarByID - No changes needed, returns car with photo URL
func GetCarByID(carID string) (*models.Car, error) {
	objID, err := primitive.ObjectIDFromHex(carID)
	if err != nil {
		return nil, errors.New("invalid car ID")
	}
	return repositories.GetCarByID(objID)
}

// validateCarImageFile checks if the uploaded car image file is valid
func validateCarImageFile(photoFile *CarPhotoFile) error {
	// Check file size (10MB limit for car photos - might be higher quality)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if photoFile.Size > maxSize {
		return errors.New("car photo file size too large (max 10MB)")
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
		return errors.New("invalid car photo file type (allowed: jpg, jpeg, png, gif, webp)")
	}

	return nil
}

// generateCarPhotoFilename creates a unique filename for car photos
func generateCarPhotoFilename(userID, plate, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	
	// Clean plate number (remove spaces, special chars for filename)
	cleanPlate := strings.ReplaceAll(strings.ReplaceAll(plate, " ", "_"), "-", "_")
	
	return fmt.Sprintf("car_%s_%s_%d%s", userID, cleanPlate, timestamp, ext)
}

// extractCarPhotoPublicIDFromURL extracts Cloudinary public_id from car photo URL
func extractCarPhotoPublicIDFromURL(url string) string {
	// Cloudinary URLs typically look like:
	// https://res.cloudinary.com/your_cloud/image/upload/v1234567890/car_photos/car_123_ABC123_1234567890.jpg
	// We need to extract: car_photos/car_123_ABC123_1234567890
	
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