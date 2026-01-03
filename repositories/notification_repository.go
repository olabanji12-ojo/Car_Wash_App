package repositories

import (
	"context"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateNotification creates a new notification in the database
func CreateNotification(notification models.Notification) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notification.ID = primitive.NewObjectID()
	notification.CreatedAt = time.Now()
	notification.IsRead = false
	notification.EmailSent = false

	result, err := database.NotificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return nil, err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)
	return &notification, nil
}

// GetNotificationsByUserID gets all notifications for a specific user
func GetNotificationsByUserID(userID string, limit int) ([]models.Notification, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	opts := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := database.NotificationCollection.Find(ctx, bson.M{"user_id": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	for cursor.Next(ctx) {
		var notification models.Notification
		if err := cursor.Decode(&notification); err == nil {
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.NotificationCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"is_read": true}},
	)

	return err
}

// MarkNotificationEmailSent marks a notification as email sent
func MarkNotificationEmailSent(notificationID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.NotificationCollection.UpdateOne(
		ctx,
		bson.M{"_id": notificationID},
		bson.M{"$set": bson.M{"email_sent": true}},
	)

	return err
}

// GetUnreadNotificationCount gets count of unread notifications for a user
func GetUnreadNotificationCount(userID string) (int64, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := database.NotificationCollection.CountDocuments(ctx, bson.M{
		"user_id": objID,
		"is_read": false,
	})

	return count, err
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func MarkAllNotificationsAsRead(userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.NotificationCollection.UpdateMany(
		ctx,
		bson.M{"user_id": objID, "is_read": false},
		bson.M{"$set": bson.M{"is_read": true}},
	)

	return err
}
