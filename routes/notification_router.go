package routes

import (
     
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

)

// NotificationRoutes sets up all routes for notification-related actions
func NotificationRoutes(router *mux.Router) {
	notifications := router.PathPrefix("/api/notifications").Subrouter()

	// All notification routes require authentication
	notifications.Use(middleware.AuthMiddleware)
    
	// User notification routes
	notifications.HandleFunc("", controllers.GetUserNotifications).Methods("GET")                    // Get user notifications
	notifications.HandleFunc("/unread-count", controllers.GetUnreadNotificationCount).Methods("GET") // Get unread count
	notifications.HandleFunc("/{id}/read", controllers.MarkNotificationAsRead).Methods("PUT")        // Mark specific as read
	notifications.HandleFunc("/mark-all-read", controllers.MarkAllNotificationsAsRead).Methods("PUT") // Mark all as read
    
	// Development/testing route
	notifications.HandleFunc("/test", controllers.TestNotification).Methods("POST") // Send test notification

}
