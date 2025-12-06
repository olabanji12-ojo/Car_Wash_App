package services

import (

	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/sirupsen/logrus"



)

type CloudinaryUploadResult struct {
	PublicID  string `json:"public_id"`
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
}

// UploadImage uploads an image to Cloudinary
func UploadImage(file multipart.File, filename string, folder string) (*CloudinaryUploadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Upload parameters
	uploadParams := uploader.UploadParams{

		Folder:    folder,        // e.g., "profile_photos"
		PublicID:  filename,      // Optional: custom public ID
		// Overwrite: true,            // Allow overwriting existing files
	 	ResourceType: "image",    // Specify resource type
		Transformation: "c_fill,h_400,w_400", // Optional: resize image
    
	}

	// Perform upload
	result, err := config.CloudinaryInstance.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		logrus.Error("Cloudinary upload failed: ", err)
		return nil, fmt.Errorf("failed to upload image: %v", err)
	}

	return &CloudinaryUploadResult{
		PublicID:  result.PublicID,
		URL:       result.URL,
		SecureURL: result.SecureURL,
	}, nil
}

// DeleteImage deletes an image from Cloudinary
func DeleteImage(publicID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := config.CloudinaryInstance.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	
	if err != nil {
		logrus.Error("Cloudinary deletion failed: ", err)
		return fmt.Errorf("failed to delete image: %v", err)
	}

	return nil
}