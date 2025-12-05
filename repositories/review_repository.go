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
)

type ReviewRepository struct {
	db *mongo.Database
}

func NewReviewRepository(db *mongo.Database) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// 1. CreateReview inserts a new review
func (rr *ReviewRepository) CreateReview(review *models.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.ReviewCollection.InsertOne(ctx, review)
	return err
}

// 2. GetReviewsByUserID returns reviews written by a specific user
func (rr *ReviewRepository) GetReviewsByUserID(userID primitive.ObjectID) ([]models.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	cursor, err := database.ReviewCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []models.Review
	for cursor.Next(ctx) {
		var review models.Review
		if err := cursor.Decode(&review); err == nil {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

// 3. GetReviewsByCarwashID returns reviews for a carwash with customer names populated
func (rr *ReviewRepository) GetReviewsByCarwashID(carwashID primitive.ObjectID) ([]models.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use aggregation pipeline to join with users collection
	pipeline := bson.A{
		// Match reviews for this carwash
		bson.M{"$match": bson.M{"carwash_id": carwashID}},
		// Lookup user details
		bson.M{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user_info",
			},
		},
		// Add customer_name field from the joined user data
		bson.M{
			"$addFields": bson.M{
				"customer_name": bson.M{
					"$arrayElemAt": bson.A{"$user_info.name", 0},
				},
			},
		},
		// Remove the temporary user_info array
		bson.M{
			"$project": bson.M{
				"user_info": 0,
			},
		},
		// Sort by created_at descending (newest first)
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
	}

	cursor, err := database.ReviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []models.Review
	for cursor.Next(ctx) {
		var review models.Review
		if err := cursor.Decode(&review); err == nil {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

// 4. GetReviewByOrderID ensures one review per order
func (rr *ReviewRepository) GetReviewByOrderID(orderID primitive.ObjectID) (*models.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var review models.Review
	err := database.ReviewCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&review)
	if err != nil {
		return nil, errors.New("no review found for this order")
	}
	return &review, nil
}

// 5. GetAverageRatingForCarwash calculates average rating
func (rr *ReviewRepository) GetAverageRatingForCarwash(carwashID primitive.ObjectID) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := bson.A{
		bson.M{"$match": bson.M{"carwash_id": carwashID}},
		bson.M{"$group": bson.M{
			"_id": "$carwash_id",
			"avg": bson.M{"$avg": "$rating"},
		}},
	}

	cursor, err := database.ReviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0, errors.New("no reviews or error calculating average")
	}

	avg, ok := result[0]["avg"].(float64)
	if !ok {
		return 0, errors.New("invalid average rating format")
	}
	return avg, nil
}

func (rr *ReviewRepository) HasUserReviewedOrder(userID, orderID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":  userID,
		"order_id": orderID,
	}

	count, err := database.ReviewCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 7. AddResponse adds a business response to a review
func (rr *ReviewRepository) AddResponse(reviewID primitive.ObjectID, response string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"response":      response,
			"response_date": time.Now(),
			"updated_at":    time.Now(),
		},
	}

	result, err := database.ReviewCollection.UpdateOne(ctx, bson.M{"_id": reviewID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("review not found")
	}

	return nil
}
