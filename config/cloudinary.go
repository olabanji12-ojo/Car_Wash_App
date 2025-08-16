package config

import (
	"github.com/cloudinary/cloudinary-go/v2"
	// "github.com/cloudinary/cloudinary-go/v2/config"
	"os"
)

var CloudinaryInstance *cloudinary.Cloudinary

func InitCloudinary() {
	cld, _ := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	CloudinaryInstance = cld
}
