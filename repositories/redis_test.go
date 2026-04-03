package repositories

import (
	"context"
	"testing"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRedisGeoIndex(t *testing.T) {
	// 1. Initialize Redis (Ensure you have Redis running locally for this test)
	// We use a separate DB or prefix if needed, but for MVP test let's just test connectivity
	client := database.ConnectRedis()
	if client == nil {
		t.Skip("Redis not available, skipping test")
	}

	repo := NewCarWashRepository(nil) // nil DB for now, we just test the SyncCarwashToRedis method

	// 2. Create a dummy carwash
	id := primitive.NewObjectID()
	carwash := &models.Carwash{
		ID: id,
		Location: models.GeoLocation{
			Type:        "Point",
			Coordinates: []float64{3.3792, 6.5244}, // Lagos, Nigeria
		},
	}

	// 3. Sync to Redis
	repo.SyncCarwashToRedis(carwash)

	// 4. Search for it
	ctx := context.Background()
	ids, err := database.RedisClient.GeoSearch(ctx, RedisGeoKey, &redis.GeoSearchQuery{
		Longitude:  3.3790, // Slightly offset
		Latitude:   6.5240,
		Radius:     5,
		RadiusUnit: "km",
	}).Result()

	if err != nil {
		t.Fatalf("Redis Search failed: %v", err)
	}

	found := false
	for _, res := range ids {
		if res == id.Hex() {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find carwash %s in Redis, but it was missing", id.Hex())
	}
}
