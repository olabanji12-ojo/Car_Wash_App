// models/geo.go
package models

// GeoPoint represents a geographic point with latitude and longitude
// Follows GeoJSON format: [longitude, latitude]
type GeoPoint struct {
    // Type should always be "Point" for a single point
    Type        string    `bson:"type" json:"-"`
    // Coordinates in [longitude, latitude] order as per GeoJSON spec
    Coordinates []float64 `bson:"coordinates" json:"-"`
}

// NewGeoPoint creates a new GeoPoint with the given coordinates
func NewGeoPoint(longitude, latitude float64) GeoPoint {
    return GeoPoint{
        Type:        "Point",
        Coordinates: []float64{longitude, latitude},
    }
}

// Longitude returns the longitude (x-coordinate)
func (g GeoPoint) Longitude() float64 {
    if len(g.Coordinates) >= 1 {
        return g.Coordinates[0]
    }
    return 0
}

// Latitude returns the latitude (y-coordinate)
func (g GeoPoint) Latitude() float64 {
    if len(g.Coordinates) >= 2 {
        return g.Coordinates[1]
    }
    return 0
}