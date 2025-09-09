package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/sirupsen/logrus"
	"fmt"
)

type CarWashRepository struct {
	db *mongo.Database
}

func NewCarWashRepository(db *mongo.Database) *CarWashRepository {
	return &CarWashRepository{db: db}
}

// 1. Create a new carwash
func (cw *CarWashRepository) CreateCarwash(carwash models.Carwash) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := database.CarwashCollection.InsertOne(ctx, carwash)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 2. Get a carwash by ID
func (cw *CarWashRepository) GetCarwashByID(objID primitive.ObjectID) (*models.Carwash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var carwash models.Carwash
	err := database.CarwashCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&carwash)
	if err != nil {
		return nil, err
	}

	return &carwash, nil
}

// 3. Get all active carwashes (for customers to browse)
func (cw *CarWashRepository) GetActiveCarwashes() ([]models.Carwash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var carwashes []models.Carwash
	for cursor.Next(ctx) {
		var cw models.Carwash
		if err := cursor.Decode(&cw); err != nil {
			logrus.Info("Error decoding carwash:", err)
		}
		carwashes = append(carwashes, cw)
	}
	fmt.Println("Carwashes found:", len(carwashes))
	return carwashes, nil
}

// 4. Update a carwash
func (cw *CarWashRepository) UpdateCarwash(id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid carwash ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	return err
}




// 5. Toggle carwash online status
func (cw *CarWashRepository) SetCarwashStatus(id string, isActive bool) error {
	return cw.UpdateCarwash(id, bson.M{"is_active": isActive})
}


func (cw *CarWashRepository) Postboarding(id primitive.ObjectID, hasOnboarded bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"has_onboarded": hasOnboarded}},
	)
	return err
}



// 6. Get all carwashes by business owner ID
func (cw *CarWashRepository) GetCarwashesByOwnerID(ownerID string) ([]models.Carwash, error) {
	objID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, bson.M{"owner_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Carwash
	for cursor.Next(ctx) {
		var carwash models.Carwash
		if err := cursor.Decode(&carwash); err == nil {
			results = append(results, carwash)
		}
	}
	return results, nil
}

// 7. Filter carwashes (optional advanced search)
func (cw *CarWashRepository) GetCarwashesByFilter(filter bson.M) ([]models.Carwash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CarwashCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var carwashes []models.Carwash
	for cursor.Next(ctx) {
		var c models.Carwash
		if err := cursor.Decode(&c); err == nil {
			carwashes = append(carwashes, c)
		}
	}
	return carwashes, nil
}

// 8. Update queue count
func (cw *CarWashRepository) UpdateQueueCount(id string, count int) error {
	return cw.UpdateCarwash(id, bson.M{"queue_count": count})
}

// 9. Find carwashes with intelligent fallback
func (cw *CarWashRepository) FindCarwashesWithFallback(userLat, userLng float64) ([]*models.Carwash, string, error) {
	// Step 1: Try 10km radius
	carwashes, err := cw.findCarwashesInRadius(userLat, userLng, 10.0)
	if err != nil {
		return nil, "", err
	}
	if len(carwashes) > 0 {
		return carwashes, "nearby", nil
	}

	// Step 2: Try 100km radius
	carwashes, err = cw.findCarwashesInRadius(userLat, userLng, 100.0)
	if err != nil {
		return nil, "", err
	}
	if len(carwashes) > 0 {
		return carwashes, "extended", nil
	}

	// Step 3: Show all with location
	carwashes, err = cw.findAllCarwashesWithLocation()
	if err != nil {
		return nil, "", err
	}
	return carwashes, "all", nil
}

// Helper function using your existing pattern
func (cw *CarWashRepository) findCarwashesInRadius(userLat, userLng, radiusKm float64) ([]*models.Carwash, error) {
	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{userLng, userLat},
				},
				"$maxDistance": radiusKm * 1000,
			},
		},
		"is_active":    true,
		"has_location": true,
	}

	carwashes, err := cw.GetCarwashesByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Convert []models.Carwash to []*models.Carwash
	var result []*models.Carwash
	for i := range carwashes {
		result = append(result, &carwashes[i])
	}
	return result, nil
}

// Helper function to get all carwashes that have location set
func (cw *CarWashRepository) findAllCarwashesWithLocation() ([]*models.Carwash, error) {
	filter := bson.M{
		"is_active":    true,
		"has_location": true,
	}

	carwashes, err := cw.GetCarwashesByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Convert []models.Carwash to []*models.Carwash
	var result []*models.Carwash
	for i := range carwashes {
		result = append(result, &carwashes[i])
	}
	return result, nil
}

// CreateService appends a new service to the carwash's services array
func (cw *CarWashRepository) CreateService(carwashID string, service models.Service) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Error("Invalid carwash ID format: ", err)
		return nil, errors.New("invalid carwash ID format")
	}

	// Validate service
	if err := service.Validate(); err != nil {
		logrus.Error("Service validation failed: ", err)
		return nil, err
	}

	// Generate new ID for the service
	service.ID = primitive.NewObjectID()

	result, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$push": bson.M{"services": service}},
	)
	if err != nil {
		logrus.Error("Failed to create service: ", err)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("carwash not found")
	}

	return result, nil
}


// GetServices retrieves the services array for a carwash
func (cw *CarWashRepository) GetServices(carwashID string) ([]models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Error("Invalid carwash ID format: ", err)
		return nil, errors.New("invalid carwash ID format")
	}

	var carwash models.Carwash
	err = database.CarwashCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&carwash)
	if err != nil {
		logrus.Error("Carwash not found: ", err)
		return nil, errors.New("carwash not found")
	}

	return carwash.Services, nil
}

// UpdateService updates a specific service in the carwash's services array
func (cw *CarWashRepository) UpdateService(carwashID, serviceID string, service models.Service) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Error("Invalid carwash ID format: ", err)
		return nil, errors.New("invalid carwash ID format")
	}

	serviceObjID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		logrus.Error("Invalid service ID format: ", err)
		return nil, errors.New("invalid service ID format")
	}

	// Validate service
	if err := service.Validate(); err != nil {
		logrus.Error("Service validation failed: ", err)
		return nil, err
	}

	// Preserve the original service ID
	service.ID = serviceObjID

	update := bson.M{
		"$set": bson.M{
			"services.$[elem].name":        service.Name,
			"services.$[elem].description": service.Description,
			"services.$[elem].price":       service.Price,
			"services.$[elem].duration":    service.Duration,
		},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": serviceObjID}},
	})

	result, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": carwashObjID},
		update,
		arrayFilters,
	)
	if err != nil {
		logrus.Error("Failed to update service: ", err)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("carwash or service not found")
	}

	return result, nil
}

// DeleteService removes a specific service from the carwash's services array
func (cw *CarWashRepository) DeleteService(carwashID, serviceID string) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		logrus.Error("Invalid carwash ID format: ", err)
		return nil, errors.New("invalid carwash ID format")
	}

	serviceObjID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		logrus.Error("Invalid service ID format: ", err)
		return nil, errors.New("invalid service ID format")
	}

	result, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": carwashObjID},
		bson.M{"$pull": bson.M{"services": bson.M{"_id": serviceObjID}}},
	)
	if err != nil {
		logrus.Error("Failed to delete service: ", err)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("carwash not found")
	}

	if result.ModifiedCount == 0 {
		return nil, errors.New("service not found")
	}

	return result, nil
}