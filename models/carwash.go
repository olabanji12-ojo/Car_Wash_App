package models

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GeoLocation struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type TimeRange struct {
	Start string `bson:"start" json:"start"`
	End   string `bson:"end" json:"end"`
}

type Service struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Duration    int                `bson:"duration" json:"duration"`
}

type Carwash struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID             primitive.ObjectID   `bson:"owner_id" json:"owner_id"` // Link to User who owns this
	Name                string               `bson:"name" json:"name"`
	Description         string               `bson:"description,omitempty" json:"description,omitempty"`
	Address             string               `bson:"address" json:"address"`
	Location            GeoLocation          `bson:"location" json:"location"`
	PhotoGallery        []string             `bson:"photo_gallery,omitempty" json:"photo_gallery,omitempty"`
	Services            []Service            `bson:"services" json:"services"`
	IsActive            bool                 `bson:"is_active" json:"is_active"`
	Rating              float64              `bson:"rating" json:"rating"`
	QueueCount          int                  `bson:"queue_count" json:"queue_count"`
	OpenHours           map[string]TimeRange `bson:"open_hours" json:"open_hours"`
	HomeService         bool                 `bson:"home_service,omitempty" json:"home_service,omitempty"`
	DeliveryRadiusKM    int                  `bson:"delivery_radius_km,omitempty" json:"delivery_radius_km,omitempty"`
	MaxCarsPerSlot      int                  `bson:"max_cars_per_slot" json:"max_cars_per_slot"`
	CreatedAt           time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `bson:"updated_at" json:"updated_at"`
	State               string               `bson:"state,omitempty" json:"state,omitempty"`
	Country             string               `bson:"country,omitempty" json:"country,omitempty"`
	LGA                 string               `bson:"lga,omitempty" json:"lga,omitempty"`
	HasLocation         bool                 `bson:"has_location" json:"has_location"`
	ServiceRangeMinutes int                  `bson:"service_range_minutes,omitempty" json:"service_range_minutes,omitempty"`
	HasOnboarded        bool                 `bson:"has_onboarded" json:"has_onboarded"`
}

func (c *Carwash) SetDefaults() {
	c.CreatedAt = time.Now()
	c.ID = primitive.NewObjectID()
	c.UpdatedAt = time.Now()
	c.QueueCount = 0
	c.IsActive = true
	c.Services = []Service{} // initialize Services as empty slice
}

func (s Service) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&s.Price, validation.Required, validation.Min(0.0)),
		validation.Field(&s.Duration, validation.Required, validation.Min(1)),
	)
}

func (t TimeRange) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Start, validation.Required),
		validation.Field(&t.End, validation.Required),
		validation.Field(&t.End, validation.By(func(value interface{}) error {
			end := value.(string)
			start := t.Start
			startTime, err := time.Parse("15:04", start)
			if err != nil {
				return err
			}
			endTime, err := time.Parse("15:04", end)
			if err != nil {
				return err
			}
			if !endTime.After(startTime) {
				return errors.New("end time must be after start time")
			}
			return nil
		})),
	)
}

func (c Carwash) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.Length(2, 100)),
		validation.Field(&c.Address, validation.Required, validation.Length(5, 200)),
		validation.Field(&c.OpenHours, validation.Required, validation.Each(validation.By(func(value interface{}) error {
			tr, ok := value.(TimeRange)
			if !ok {
				return errors.New("invalid time range")
			}
			return tr.Validate()
		}))),
		validation.Field(&c.DeliveryRadiusKM, validation.When(c.HomeService, validation.Min(1))),
	)
}

// GetDistanceFrom calculates the distance between the carwash and a given location
// Returns distance in kilometers and a human-readable string
func (c *Carwash) GetDistanceFrom(userLat, userLng float64) (distanceKm float64, distanceText string) {
	if len(c.Location.Coordinates) != 2 {
		return 0, "N/A"
	}

	// Get carwash coordinates (note: GeoJSON uses [longitude, latitude] order)
	lng, lat := c.Location.Coordinates[0], c.Location.Coordinates[1]

	// Calculate distance using your existing utility
	distanceKm = utils.CalculateDistance(userLat, userLng, lat, lng)

	// Format distance text
	if distanceKm < 1 {
		distanceText = fmt.Sprintf("%.0f m away", distanceKm*1000)
	} else {
		distanceText = fmt.Sprintf("%.1f km away", distanceKm)
	}

	return distanceKm, distanceText
}

// WithDistance returns a map with carwash details including distance information
// This is useful for API responses
func (c *Carwash) WithDistance(userLat, userLng float64) map[string]interface{} {
	distanceKm, distanceText := c.GetDistanceFrom(userLat, userLng)

	return map[string]interface{}{
		"id":                 c.ID,
		"name":               c.Name,
		"address":            c.Address,
		"location":           c.Location,
		"rating":             c.Rating,
		"queue_count":        c.QueueCount,
		"is_active":          c.IsActive,
		"home_service":       c.HomeService,
		"delivery_radius_km": c.DeliveryRadiusKM,
		"photo":              c.getMainPhoto(),
		"photo_gallery":      c.PhotoGallery,
		"distance_km":        distanceKm,
		"distance_text":      distanceText,
		"state":              c.State,
		"lga":                c.LGA,
		"country":            c.Country,
	}
}

// getMainPhoto returns the first photo from the gallery or an empty string
func (c *Carwash) getMainPhoto() string {
	if len(c.PhotoGallery) > 0 {
		return c.PhotoGallery[0]
	}
	return ""
}
