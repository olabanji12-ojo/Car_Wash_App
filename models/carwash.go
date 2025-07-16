package models

import (

	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

)

// Embedded struct for map location
type GeoLocation struct {

	Type        string    `bson:"type" json:"type"` // always "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [lng, lat]
}


type Carwash struct {

	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID       primitive.ObjectID   `bson:"owner_id" json:"owner_id"`                         // Linked to users._id
	Name          string               `bson:"name" json:"name"`                                 // Business name
	Description   string               `bson:"description,omitempty" json:"description,omitempty"` // Optional business info
	Address       string               `bson:"address" json:"address"`                           // Full address
	Location      GeoLocation          `bson:"location" json:"location"`                         // Geo search
	PhotoGallery  []string             `bson:"photo_gallery,omitempty" json:"photo_gallery,omitempty"` // Images
	Services      []primitive.ObjectID `bson:"services" json:"services"`                         // List of service IDs
	IsActive      bool                 `bson:"is_active" json:"is_active"`                       // Can accept bookings?
	// Rating        float64              `bson:"rating" json:"rating"`                             // Avg from reviews
	QueueCount    int                  `bson:"queue_count" json:"queue_count"`                   // Cars waiting
	OpenHours     map[string]string    `bson:"open_hours" json:"open_hours"`                     // E.g., { "mon": "8am–6pm" }
	HomeService   bool                 `bson:"home_service" json:"home_service"`                 // Mobile service?
	DeliveryRadiusKM int               `bson:"delivery_radius_km,omitempty" json:"delivery_radius_km,omitempty"` // e.g., 10 means 10km max
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`                     // Joined on
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`                     // Last update
	

}




