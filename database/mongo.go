package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() *mongo.Database {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using defaults")
	}

	mongoURL := os.Getenv("MONGO_URL")
	

	dbName := os.Getenv("DB_NAME")
	
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("❌ Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	fmt.Println("✅ MongoDB connected successfully")
	return DB
}
