package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	// validation "github.com/go-ozzo/ozzo-validation/v4"
	// "errors"
	// "fmt"
)

// Embedded struct for map location
type GeoLocation struct {
	Type        string    `bson:"type" json:"type"`               // always "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [lng, lat]
}

// Time Range Opening for the Car wash

type TimeRange struct {
	Start string `bson:"start" json:"start"` // "08:00"
	End   string `bson:"end" json:"end"`     // "18:00"
}


type Carwash struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`                                   // Business name
	Description string             `bson:"description,omitempty" json:"description,omitempty"` // Optional business info
	Address     string             `bson:"address" json:"address"`                             // Full address
	Location    GeoLocation        `bson:"location" json:"location"`                           // Geo search

	PhotoGallery     []string             `bson:"photo_gallery,omitempty" json:"photo_gallery,omitempty"`           // Images
	Services         []Service            `bson:"services" json:"services"`                                         // List of service IDs
	IsActive         bool                 `bson:"is_active" json:"is_active"`                                       // Can accept bookings?
	Rating           float64              `bson:"rating" json:"rating"`                                             // Avg from reviews
	QueueCount       int                  `bson:"queue_count" json:"queue_count"`                                   // Cars waiting
	OpenHours        map[string]TimeRange `bson:"open_hours" json:"open_hours"`                                     // E.g., { "mon": "8am–6pm" }
	HomeService      bool                 `bson:"home_service,omitempty" json:"home_service,omitempty"`             // Mobile service?
	DeliveryRadiusKM int                  `bson:"delivery_radius_km,omitempty" json:"delivery_radius_km,omitempty"` // e.g., 10 means 10km max
	CreatedAt        time.Time            `bson:"created_at" json:"created_at"`                                     // Joined on
	UpdatedAt        time.Time            `bson:"updated_at" json:"updated_at"`                                     // Last update
	State            string               `bson:"state,omitempty" json:"state,omitempty"`
	Country          string               `bson:"country,omitempty" json:"country,omitempty"`
	LGA              string               `bson:"lga,omitempty" json:"lga,omitempty"`
	HasLocation      bool                 `bson:"has_location" json:"has_location"` // true if location is set

	ServiceRangeMinutes int `bson:"service_range_minutes,omitempty" json:"service_range_minutes,omitempty"`
}

type Service struct {
	Name        string  `bson:"name" json:"name"`
	Description string  `bson:"description" json:"description"`
	Price       float64 `bson:"price" json:"price"`
	Duration    int     `bson:"duration" json:"duration"`
}

func (c *Carwash) SetDefaults() {
	c.CreatedAt = time.Now()
	c.ID = primitive.NewObjectID()
	c.UpdatedAt = time.Now()
	c.QueueCount = 0
	c.IsActive = true
    
}

// func (c Carwash) Validate() error {
// 	return validation.ValidateStruct(&c,
// 		validation.Field(&c.Name, validation.Required, validation.Length(3, 100)),
// 		validation.Field(&c.Address, validation.Required, validation.Length(5, 200)),
// 		validation.Field(&c.Location.Type, validation.Required, validation.In("Point")),
// 		validation.Field(&c.Location.Coordinates, validation.Required, validation.Length(2, 2)),
// 		// validation.Field(&c.OpenHours, validation.By(validateOpenHours)),
// 	)
// }

// func validateOpenHours(value interface{}) error {
// 	hours, ok := value.(map[string]TimeRange)
// 	if !ok {
// 		return errors.New("invalid open_hours format")
// 	}
// 	for day, t := range hours {
// 		if t.Start == "" || t.End == "" {
// 			return fmt.Errorf("missing start or end time for %s", day)
// 		}
// 	}
// 	return nil
// }
