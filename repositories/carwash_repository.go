package repositories

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const RedisGeoKey = "carwashes:geo"

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

	// Sync to Redis (Fail-soft)
	go cw.SyncCarwashToRedis(&carwash)

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

	// If location was updated, sync to Redis
	if _, ok := update["location"]; ok {
		if carwash, err := cw.GetCarwashByID(objID); err == nil {
			go cw.SyncCarwashToRedis(carwash)
		}
	}

	return err
}

// 5. Toggle carwash online status
func (cw *CarWashRepository) SetCarwashStatus(id string, isActive bool) error {
	return cw.UpdateCarwash(id, bson.M{"is_active": isActive})
}

// 6. Post onboarding status
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

// 7. Get all carwashes by business owner ID
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

// 8. Filter carwashes (optional advanced search)
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

// 9. RefreshQueueCount - recalculates queue count from active orders and updates the carwash record
func (cw *CarWashRepository) RefreshQueueCount(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid carwash ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Count active orders for this carwash
	count, err := database.OrderCollection.CountDocuments(ctx, bson.M{
		"carwash_id": objID,
		"status":     "active",
	})
	if err != nil {
		return fmt.Errorf("failed to recalculate queue count: %w", err)
	}

	// Update the carwash document
	_, err = database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"queue_count": int(count), "updated_at": time.Now()}},
	)
	return err
}

// 10. UpdateCarwashRating - updates the average rating and review count for a carwash
func (cw *CarWashRepository) UpdateCarwashRating(id primitive.ObjectID, averageRating float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"rating":     averageRating,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

// 10. Find carwashes with intelligent fallback
func (cw *CarWashRepository) FindCarwashesWithFallback(userLat, userLng float64) ([]map[string]interface{}, string, error) {
	// Step 1: Try 10km radius
	carwashes, err := cw.findCarwashesInRadius(userLat, userLng, 10.0)
	if err != nil {
		return nil, "", err
	}
	if len(carwashes) > 0 {
		result := cw.addDistanceToCarwashes(carwashes, userLat, userLng)
		return result, "nearby", nil
	}

	// Step 2: Try 100km radius
	carwashes, err = cw.findCarwashesInRadius(userLat, userLng, 100.0)
	if err != nil {
		return nil, "", err
	}
	if len(carwashes) > 0 {
		result := cw.addDistanceToCarwashes(carwashes, userLat, userLng)
		return result, "extended", nil
	}

	// Step 3: Show all with location
	carwashes, err = cw.findAllCarwashesWithLocation()
	if err != nil {
		return nil, "", err
	}

	result := cw.addDistanceToCarwashes(carwashes, userLat, userLng)
	return result, "all", nil
}

// Helper function using your existing pattern (Optimized with Redis)
func (cw *CarWashRepository) findCarwashesInRadius(userLat, userLng, radiusKm float64) ([]*models.Carwash, error) {
	// Step 1: Try Redis first
	if database.RedisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Redis GeoSearch (v9)
		ids, err := database.RedisClient.GeoSearch(ctx, RedisGeoKey, &redis.GeoSearchQuery{
			Member:     "", // Not searching relative to another member
			Longitude:  userLng,
			Latitude:   userLat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "asc",
		}).Result()

		if err == nil && len(ids) > 0 {
			logrus.Infof("🚀 Redis Geo-Hit: Found %d shops within %.1fkm", len(ids), radiusKm)

			var objectIDs []primitive.ObjectID
			for _, idStr := range ids {
				if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
					objectIDs = append(objectIDs, oid)
				}
			}

			// Fetch full details from Mongo, ensuring we respect status
			filter := bson.M{
				"_id":       bson.M{"$in": objectIDs},
				"is_active": true,
			}
			carwashes, err := cw.GetCarwashesByFilter(filter)
			if err == nil && len(carwashes) > 0 {
				var result []*models.Carwash
				// Note: Mongo $in doesn't preserve order. 
				// For the MVP, we'll let the frontend sort by distance or add distance later.
				for i := range carwashes {
					result = append(result, &carwashes[i])
				}
				return result, nil
			}
		}
	}

	// Step 2: Fallback to MongoDB $near
	logrus.Info("⚠️ Redis Miss or Error: Falling back to MongoDB $near")
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

// Helper function to add distance information to carwashes
func (cw *CarWashRepository) addDistanceToCarwashes(carwashes []*models.Carwash, userLat, userLng float64) []map[string]interface{} {
	var result []map[string]interface{}
	for _, cw := range carwashes {
		if cw != nil {
			result = append(result, cw.WithDistance(userLat, userLng))
		}
	}

	// Sort by distance (nearest first)
	sort.Slice(result, func(i, j int) bool {
		distI, okI := result[i]["distance_km"].(float64)
		distJ, okJ := result[j]["distance_km"].(float64)

		if !okI || !okJ {
			return false
		}
		return distI < distJ
	})

	return result
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

// AddPhotoToGallery adds a photo URL to the carwash's photo gallery
func (cw *CarWashRepository) AddPhotoToGallery(id primitive.ObjectID, photoURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CarwashCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$push": bson.M{"photo_gallery": photoURL}},
	)
	return err
}

// SyncCarwashToRedis - helper to sync a carwash's location to the Redis Geo Index
func (cw *CarWashRepository) SyncCarwashToRedis(carwash *models.Carwash) {
	if database.RedisClient == nil || carwash.Location.Coordinates == nil || len(carwash.Location.Coordinates) < 2 {
		logrus.Warnf("Skipping Redis Sync for Carwash %s: Redis client or location is missing", carwash.ID.Hex())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Redis GEOADD uses [longitude, latitude]
	err := database.RedisClient.GeoAdd(ctx, RedisGeoKey, &redis.GeoLocation{
		Name:      carwash.ID.Hex(),
		Longitude: carwash.Location.Coordinates[0],
		Latitude:  carwash.Location.Coordinates[1],
	}).Err()

	if err != nil {
		logrus.Errorf("❌ Failed to sync carwash %s to Redis: %v", carwash.ID.Hex(), err)
	} else {
		logrus.Infof("✅ Carwash %s synced to Redis Geo Index", carwash.ID.Hex())
	}
}
