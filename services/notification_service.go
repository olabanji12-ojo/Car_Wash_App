package services

import (
	"fmt"
	"log"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService handles all notification operations
type NotificationService struct{}

// CreateNotification creates a notification and optionally sends email
func (ns *NotificationService) CreateNotification(userID primitive.ObjectID, title, message, notificationType string, sendEmail bool) error {
	notification := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notificationType,
	}

	// Validate notification
	if err := notification.Validate(); err != nil {
		return fmt.Errorf("notification validation failed: %v", err)
	}

	// Save to database
	savedNotification, err := repositories.CreateNotification(notification)
	if err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send email if requested (async to avoid blocking)
	if sendEmail {
		go ns.sendEmailAsync(savedNotification.ID, userID, title, message)
	}

	return nil
}

// sendEmailAsync sends email in background (like Django's async tasks)
func (ns *NotificationService) sendEmailAsync(notificationID, userID primitive.ObjectID, title, message string) {
	// Get user email
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user for email notification: %v", err)
		return
	}

	// Send email
	err = utils.SendEmail(user.Email, title, message)
	if err != nil {
		log.Printf("Failed to send email notification: %v", err)
		return
	}

	// Mark email as sent
	err = repositories.MarkNotificationEmailSent(notificationID)
	if err != nil {
		log.Printf("Failed to mark email as sent: %v", err)
	}

	log.Printf("Email notification sent successfully to %s", user.Email)
}

// BOOKING NOTIFICATION TRIGGERS (Like Django Signals)

// SendBookingConfirmation - triggered when booking is created
func (ns *NotificationService) SendBookingConfirmation(booking *models.Booking) {
	title := "Booking Confirmation"
	message := fmt.Sprintf("Your carwash booking has been confirmed for %s", booking.BookingTime.Format("Jan 2, 2006 at 3:04 PM"))
	
	err := ns.CreateNotification(booking.UserID, title, message, models.NotificationTypeBooking, true)
	if err != nil {
		log.Printf("Failed to send booking confirmation: %v", err)
	}
}

// SendBookingAccepted - triggered when business accepts booking
func (ns *NotificationService) SendBookingAccepted(booking *models.Booking, carwashName string) {
	title := "Booking Accepted!"
	message := fmt.Sprintf("Great news! %s has accepted your booking for %s", carwashName, booking.BookingTime.Format("Jan 2, 2006 at 3:04 PM"))
	
	err := ns.CreateNotification(booking.UserID, title, message, models.NotificationTypeBooking, true)
	if err != nil {
		log.Printf("Failed to send booking accepted notification: %v", err)
	}
}

// SendBookingRejected - triggered when business rejects booking
func (ns *NotificationService) SendBookingRejected(booking *models.Booking, reason string) {
	title := "Booking Update"
	message := fmt.Sprintf("Unfortunately, your booking for %s could not be confirmed. Reason: %s", booking.BookingTime.Format("Jan 2, 2006"), reason)
	
	err := ns.CreateNotification(booking.UserID, title, message, models.NotificationTypeBooking, true)
	if err != nil {
		log.Printf("Failed to send booking rejected notification: %v", err)
	}
}

// ORDER NOTIFICATION TRIGGERS

// SendOrderCreated - triggered when order is created from booking
func (ns *NotificationService) SendOrderCreated(order *models.Order) {
	title := "Order Created"
	message := "Your booking has been converted to an active order. We'll notify you when a worker is assigned."
	
	err := ns.CreateNotification(order.UserID, title, message, models.NotificationTypeOrder, true)
	if err != nil {
		log.Printf("Failed to send order created notification: %v", err)
	}
}

// SendWorkerAssigned - triggered when worker is assigned to order
func (ns *NotificationService) SendWorkerAssigned(order *models.Order, workerName string) {
	title := "Worker Assigned"
	message := fmt.Sprintf("Good news! %s has been assigned to your order and will be with you soon.", workerName)
	
	err := ns.CreateNotification(order.UserID, title, message, models.NotificationTypeWorker, true)
	if err != nil {
		log.Printf("Failed to send worker assigned notification: %v", err)
	}
}

// SendOrderStatusUpdate - triggered when order status changes
func (ns *NotificationService) SendOrderStatusUpdate(order *models.Order, newStatus, details string) {
	title := fmt.Sprintf("Order %s", newStatus)
	message := details
	if message == "" {
		message = fmt.Sprintf("Your order status has been updated to: %s", newStatus)
	}
	
	err := ns.CreateNotification(order.UserID, title, message, models.NotificationTypeOrder, true)
	if err != nil {
		log.Printf("Failed to send order status update: %v", err)
	}
}

// BUSINESS NOTIFICATION TRIGGERS

// SendNewBookingToBusiness - notify business of new booking
func (ns *NotificationService) SendNewBookingToBusiness(businessUserID primitive.ObjectID, customerName, serviceName string) {
	title := "New Booking Received"
	message := fmt.Sprintf("You have a new booking from %s for %s. Please review and accept/reject.", customerName, serviceName)
	
	err := ns.CreateNotification(businessUserID, title, message, models.NotificationTypeBooking, true)
	if err != nil {
		log.Printf("Failed to send new booking notification to business: %v", err)
	}
}

// GetUserNotifications gets notifications for a user
func (ns *NotificationService) GetUserNotifications(userID string, limit int) ([]models.Notification, error) {
	return repositories.GetNotificationsByUserID(userID, limit)
}

// MarkAsRead marks a notification as read
func (ns *NotificationService) MarkAsRead(notificationID string) error {
	return repositories.MarkNotificationAsRead(notificationID)
}

// GetUnreadCount gets unread notification count for a user
func (ns *NotificationService) GetUnreadCount(userID string) (int64, error) {
	return repositories.GetUnreadNotificationCount(userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (ns *NotificationService) MarkAllAsRead(userID string) error {
	return repositories.MarkAllNotificationsAsRead(userID)
}

// Global notification service instance
var NotificationSvc = &NotificationService{}
