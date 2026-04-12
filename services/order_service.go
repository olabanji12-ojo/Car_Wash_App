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

type OrderService struct {
	orderRepository   repositories.OrderRepository
	bookingRepository repositories.BookingRepository
	carWashRepository repositories.CarWashRepository
	financeService    *FinanceService
}

func NewOrderService(orderRepository repositories.OrderRepository, bookingRepository repositories.BookingRepository, carWashRepository repositories.CarWashRepository, financeService *FinanceService) *OrderService {
	return &OrderService{
		orderRepository:   orderRepository,
		bookingRepository: bookingRepository,
		carWashRepository: carWashRepository,
		financeService:    financeService,
	}
}


// ✅ CreateOrderFromBooking
func(os *OrderService) CreateOrderFromBooking(bookingID string) (*models.Order, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Convert ID
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	// 2. Fetch the booking
	booking, err := os.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	// 3. Prevent duplicate order
	if booking.Status == "approved" || booking.Status == "completed" {
		return nil, errors.New("booking already approved or processed")
	}

	// 4. Build the order
	newOrder := models.Order{
		ID:            primitive.NewObjectID(),
		BookingID:     booking.ID,
		UserID:        booking.UserID,
		CarID:         booking.CarID,
		CarwashID:     booking.CarwashID,
		QueueNumber:   booking.QueueNumber,
		BookingType:   booking.BookingType,
		UserLocation:  booking.UserLocation,
		Status:        "active",
		PaymentStatus: "unpaid",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 5. Calculate Total Amount from Carwash Services
	total := 0.0
	carwash, err := os.carWashRepository.GetCarwashByID(booking.CarwashID)
	if err == nil {
		for _, sID := range booking.Services {
			for _, s := range carwash.Services {
				if s.ID == sID {
					total += s.Price
				}
			}
		}
	} else {
		logrus.Errorf("❌ Failed to fetch carwash %s for pricing: %v", booking.CarwashID.Hex(), err)
	}

	// 6. Handle FREE Wash logic (Loyalty Reward)
	if booking.IsFree {
		newOrder.TotalAmount = 0
		newOrder.PlatformFee = 0
		newOrder.MerchantNet = 0
		newOrder.PaymentStatus = "paid" // It's free!
		logrus.Infof("🎁 Order %s marked as FREE due to loyalty reward", newOrder.ID.Hex())
	} else {
		newOrder.TotalAmount = total
		// 7. Calculate Splits using FinanceService (Loosely Coupled)
		pFee, mNet := os.financeService.CalculateSplit(total)
		newOrder.PlatformFee = pFee
		newOrder.MerchantNet = mNet
	}

	// 5. Save the order
	if err := os.orderRepository.CreateOrder(&newOrder); err != nil {
		logrus.Error("Failed to create order: ", err)
		return nil, err
	}

	// 6. Update booking status to approved
	if err := os.bookingRepository.UpdateBookingStatus(booking.ID, "approved"); err != nil {
		logrus.Warn("Order created, but failed to update booking status")
	}

	return &newOrder, nil
}


func(os *OrderService) GetOrderByID(orderID string) (*models.Order, error) {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID")
	}

	return os.orderRepository.GetOrderByID(objID)
}


func(os *OrderService) GetOrdersByUser(userID string) ([]models.Order, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return os.orderRepository.GetOrdersByUserID(uid)
}

func(os *OrderService) GetOrdersByCarwash(carwashID string) ([]models.Order, error) {

	owner_id, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return os.orderRepository.GetOrdersByCarwashID(owner_id)
	
} 



func (os *OrderService) UpdateOrderStatus(orderID string, newStatus string, photoURL string) error {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID")
	}

	// Fetch order to check type
	order, err := os.orderRepository.GetOrderByID(objID)
	if err != nil {
		return errors.New("order not found")
	}

	// VALIDATION: Proof of Work (Completion Photo)
	if newStatus == "completed" && order.BookingType == "home_service" {
		if photoURL == "" {
			return errors.New("cannot complete: proof-of-work photo is required for home services")
		}
		// Save the photo URL
		updates := primitive.M{
			"completion_photo_url": photoURL,
			"updated_at":           time.Now(),
		}
		os.orderRepository.UpdateOrder(objID, updates)
	}

	if err := os.orderRepository.UpdateOrderStatus(objID, newStatus); err != nil {
		return err
	}

	// Step 4: AUTO-LEDGER (If completed, record the payment)
	if newStatus == "completed" {
		// Calculate the split using FinanceService (returns fee, net)
		platformFee, merchantNet := os.financeService.CalculateSplit(order.TotalAmount)

		payment := &models.Payment{
			ID:             primitive.NewObjectID(),
			UserID:         order.UserID,
			CarwashID:      order.CarwashID,
			OrderID:        order.ID,
			Amount:         order.TotalAmount,
			PlatformFee:    platformFee,
			MerchantNet:    merchantNet,
			Method:         "cash", // Default for now
			Status:         "paid",
			PaidAt:         time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := repositories.CreatePayment(payment); err != nil {
			logrus.Errorf("❌ Failed to record payment for order %s: %v", orderID, err)
			// We don't return error here because the order IS completed, 
			// we just failed to log it (should be retried in a real world system).
		} else {
			logrus.Infof("✅ Payment ledger updated for order %s (Net: %.2f)", orderID, merchantNet)
		}
	}

	return nil
}


func(os *OrderService) AssignWorker(orderID string, workerID string) error {
	oid, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return errors.New("invalid order ID")
	}

	wid, err := primitive.ObjectIDFromHex(workerID)
	if err != nil {
		return errors.New("invalid worker ID")
	}

	return os.orderRepository.AssignWorker(oid, wid)
	
}



