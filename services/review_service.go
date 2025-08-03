package services

import (

	"context"
	"errors"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

)

//  LeaveReview allows a user to review a carwash after a completed order
func CreateReview(userID string, input models.Review) (*models.Review, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Convert IDs
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// orderID, err := primitive.ObjectIDFromHex(input.OrderID)
	
	reviewExists, err := repositories.HasUserReviewedOrder(userObjID, *input.OrderID)
	if err != nil {
		return nil, errors.New("error checking for existing review")
	}
	if reviewExists {
		return nil, errors.New("you've already submitted a review for this order")
	}


	// 4. Create Review object
	newReview := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    userObjID,
		OrderID:   input.OrderID,
		CarwashID: input.CarwashID,
		Rating:    input.Rating,
		Comment:   input.Comment,
		Photos:    input.Photos,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 5. Save to DB
	if err := repositories.CreateReview(&newReview); err != nil {
		logrus.Error("Error creating review: ", err)
		return nil, errors.New("failed to create review")
	}

	return &newReview, nil
}

//  GetReviewsByUserID fetches reviews made by a user
func GetReviewsByUserID(userID string) ([]models.Review, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return repositories.GetReviewsByUserID(userObjID)

}

//  GetReviewsByCarwashID fetches reviews for a specific carwash
func GetReviewsByCarwashID(carwashID string) ([]models.Review, error) {
	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return repositories.GetReviewsByCarwashID(carwashObjID)
}

//  GetReviewByOrderID fetches the review for a specific order
func GetReviewByOrderID(orderID string) (*models.Review, error) {
	orderObjID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID")
	}

	return repositories.GetReviewByOrderID(orderObjID)
}

//  GetAverageRating calculates the average rating of a carwash
func GetAverageRating(carwashID string) (float64, error) {
	carwashObjID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return 0, errors.New("invalid carwash ID")
	}

	return repositories.GetAverageRatingForCarwash(carwashObjID)
}




