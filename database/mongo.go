package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() *mongo.Database {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using defaults")
	}
 
	mongoURL := os.Getenv("MONGO_URI")
	log.Println(mongoURL)
	

	dbName := os.Getenv("DB_NAME")
	
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("❌ Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ MongoDB ping failed: %v", err)
	}

	DB = client.Database(dbName)
	
	// Create indexes
	if err := createIndexes(); err != nil {
		log.Fatalf("❌ Failed to create indexes: %v", err)
	}
	
	fmt.Println("✅ MongoDB connected and indexes created successfully")
	return DB
}

// createIndexes creates all necessary database indexes
func createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Geospatial index for user locations
	userLocationIndex := mongo.IndexModel{
		Keys: bson.M{"last_location": "2dsphere",
		},
	}

	// 2. Index for address lookups
	addressIndex := mongo.IndexModel{
		Keys: bson.M{"addresses._id": 1},
	}

	// 3. Index for user email (unique)
	emailIndex := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}

	// 4. Index for user phone (unique)
	phoneIndex := mongo.IndexModel{
		Keys:    bson.M{"phone": 1},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}

	// Create all indexes
	_, err := DB.Collection("users").Indexes().CreateMany(ctx, []mongo.IndexModel{
		userLocationIndex,
		addressIndex,
		emailIndex,
		phoneIndex,
	})

	if err != nil {
		return fmt.Errorf("failed to create user indexes: %v", err)
	}

	// Add indexes for other collections as needed
	// Example for carwashes collection
	carwashLocationIndex := mongo.IndexModel{
		Keys: bson.M{"location": "2dsphere"},
	}

	_, err = DB.Collection("carwashes").Indexes().CreateOne(ctx, carwashLocationIndex)
	if err != nil {
		return fmt.Errorf("failed to create carwash location index: %v", err)
	}

	return nil
}