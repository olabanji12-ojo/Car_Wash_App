package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	// "github.com/go-ozzo/ozzo-validation/v4/is"
	"errors"

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
	ID               primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name             string               `bson:"name" json:"name"`
	Description      string               `bson:"description,omitempty" json:"description,omitempty"`
	Address          string               `bson:"address" json:"address"`
	Location         GeoLocation          `bson:"location" json:"location"`
	PhotoGallery     []string             `bson:"photo_gallery,omitempty" json:"photo_gallery,omitempty"`
	Services         []Service            `bson:"services" json:"services"`
	IsActive         bool                 `bson:"is_active" json:"is_active"`
	Rating           float64              `bson:"rating" json:"rating"`
	QueueCount       int                  `bson:"queue_count" json:"queue_count"`
	OpenHours        map[string]TimeRange `bson:"open_hours" json:"open_hours"`
	HomeService      bool                 `bson:"home_service,omitempty" json:"home_service,omitempty"`
	DeliveryRadiusKM int                  `bson:"delivery_radius_km,omitempty" json:"delivery_radius_km,omitempty"`
	MaxCarsPerSlot   int                  `bson:"max_cars_per_slot" json:"max_cars_per_slot"`
	CreatedAt        time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time            `bson:"updated_at" json:"updated_at"`
	State            string               `bson:"state,omitempty" json:"state,omitempty"`
	Country          string               `bson:"country,omitempty" json:"country,omitempty"`
	LGA              string               `bson:"lga,omitempty" json:"lga,omitempty"`
	HasLocation      bool                 `bson:"has_location" json:"has_location"`
	ServiceRangeMinutes int               `bson:"service_range_minutes,omitempty" json:"service_range_minutes,omitempty"`
	HasOnboarded     bool                  `bson:"has_onboarded" json:"has_onboarded"`
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

