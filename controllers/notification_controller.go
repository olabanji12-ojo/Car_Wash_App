package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserNotifications gets notifications for the authenticated user
func GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	// Get limit from query params (default: 20)
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get notifications
	notifications, err := services.NotificationSvc.GetUserNotifications(userID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// MarkNotificationAsRead marks a specific notification as read
func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	notificationID := mux.Vars(r)["id"]

	err := services.NotificationSvc.MarkAsRead(notificationID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

// GetUnreadNotificationCount gets count of unread notifications
func GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	count, err := services.NotificationSvc.GetUnreadCount(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"unread_count": count,
	})
}

// MarkAllNotificationsAsRead marks all notifications as read for user
func MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	err := services.NotificationSvc.MarkAllAsRead(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "All notifications marked as read",
	})
}

// TestNotification sends a test notification (development only)
func TestNotification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	var payload struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	// Convert userID string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Send test notification
	err = services.NotificationSvc.CreateNotification(
		userObjID,
		payload.Title,
		payload.Message,
		"general",
		true, // Send email
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Test notification sent successfully",
	})
}
