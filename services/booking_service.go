package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"strings"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingService struct {
	bookingRepository   repositories.BookingRepository
	carWashRepository   repositories.CarWashRepository
	userRepository      repositories.UserRepository
	notificationService *NotificationService
	mapService          MapService
	loyaltyService      *LoyaltyService
	lockService         LockService
	walletService       *WalletService
	financeService      *FinanceService
}

func NewBookingService(bookingRepository repositories.BookingRepository, carWashRepository repositories.CarWashRepository, userRepository repositories.UserRepository, notificationService *NotificationService, mapService MapService, loyaltyService *LoyaltyService, lockService LockService, walletService *WalletService, financeService *FinanceService) *BookingService {
	return &BookingService{
		bookingRepository:   bookingRepository,
		carWashRepository:   carWashRepository,
		userRepository:      userRepository,
		notificationService: notificationService,
		mapService:          mapService,
		loyaltyService:      loyaltyService,
		lockService:         lockService,
		walletService:       walletService,
		financeService:      financeService,
	}
}

// CreateBooking for a selected time slot
func (bs *BookingService) CreateBooking(userID string, input models.Booking) (*models.Booking, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Convert userID to ObjectID
	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Step 2: Validate BookingTime is not empty
	if input.BookingTime.IsZero() {
		return nil, errors.New("booking time is required")
	}

	//  Step 3: FETCH the Carwash and VALIDATE against its OpenHours
	carwash, err := bs.carWashRepository.GetCarwashByID(input.CarwashID)
	if err != nil {
		return nil, errors.New("carwash not found")
	}

	// Get the weekday from booking time (use full name to match open_hours keys)
	weekday := strings.ToLower(input.BookingTime.Weekday().String()) // "monday", "tuesday", etc

	logrus.Infof("Checking open hours for carwash %s on %s (derived from %v)", input.CarwashID, weekday, input.BookingTime)
	logrus.Infof("Carwash OpenHours keys: %v", carwash.OpenHours)

	timeRange, ok := carwash.OpenHours[weekday]
	if !ok {
		logrus.Warnf("Carwash %s is not open on %s. Available days: %v", input.CarwashID, weekday, carwash.OpenHours)
		return nil, errors.New("carwash is not open on this day")
	}

	// Convert time strings to time.Time
	layout := "15:04"
	bookingHour := input.BookingTime.Format(layout)

	bookingParsed, err := time.Parse(layout, bookingHour)
	if err != nil {
		return nil, errors.New("invalid booking time format")
	}

	startParsed, err := time.Parse(layout, timeRange.Start)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}

	endParsed, err := time.Parse(layout, timeRange.End)
	if err != nil {
		return nil, errors.New("invalid end time format")
	}

	// Compare booking time with open hours
	if bookingParsed.Before(startParsed) || bookingParsed.After(endParsed) {
		return nil, errors.New("booking time is outside of open hours")
	}

	// Step 4: Check for existing bookings on same date/time
	bookingsForDay, err := bs.bookingRepository.GetBookingsByDate(input.CarwashID, input.BookingTime)
	if err != nil {
		return nil, errors.New("could not fetch bookings for that time")
	}

	maxCars := carwash.MaxCarsPerSlot
	if maxCars <= 0 {
		maxCars = 1
	}

	currentCarsInSlot := 0
	inputTimeStr := input.BookingTime.UTC().Truncate(time.Minute).Format("2006-01-02 15:04")
	logrus.Infof("[CreateDebug] User: %s, Slot: %s, Max: %d", userID, inputTimeStr, maxCars)

	for _, b := range bookingsForDay {
		bTimeStr := b.BookingTime.UTC().Truncate(time.Minute).Format("2006-01-02 15:04")

		// If same user already has a pending/confirmed booking for this slot, block them
		if b.UserID.Hex() == userID && bTimeStr == inputTimeStr && b.Status != "cancelled" && b.Status != "completed" {
			return nil, errors.New("you already have a booking for this time slot")
		}

		if b.Status != "confirmed" {
			continue
		}

		if bTimeStr == inputTimeStr {
			currentCarsInSlot++
		}
	}

	if currentCarsInSlot >= maxCars {
		return nil, errors.New("selected time slot is already fully booked")
	}

	// Step 4.5: FINAL Lock Verification (Ensure it wasn't stolen at the last millisecond)
	lockKey := GetLockKey(input.CarwashID.Hex(), inputTimeStr)
	lockedByOther, _ := bs.lockService.IsLockedByOther(lockKey, ownerID.Hex())
	if lockedByOther {
		return nil, errors.New("someone else just grabbed this slot! Please try another time.")
	}

	// Step 5: Create new booking
	queueNumber := len(bookingsForDay) + 1
	verificationCode, _ := utils.GenerateNumericCode(4)

	newBooking := models.Booking{

		ID:               primitive.NewObjectID(),
		UserID:           ownerID,
		CarID:            input.CarID,
		CarwashID:        input.CarwashID,
		BookingTime:      input.BookingTime,
		BookingType:      input.BookingType,
		UserLocation:     input.UserLocation,
		AddressNote:      input.AddressNote,
		Notes:            input.Notes,
		Status:           "pending",
		QueueNumber:      queueNumber,
		VerificationCode: verificationCode,
		Services:         input.Services,
		TotalPrice:       input.TotalPrice,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ExpiresAt:        time.Now().Add(5 * time.Minute),
	}

	// Step 6: Apply Loyalty Reward (100% discount)
	isEligible, _, _ := bs.loyaltyService.IsEligibleForFreeWash(ownerID.Hex())
	if isEligible {
		newBooking.IsFree = true
		logrus.Infof("🎁 User %s is eligible for a FREE wash (Loyalty Reward)", ownerID.Hex())
	}

	// Check distance if it's a home service
	// Step 5.5: Distance check for home service
	if input.BookingType == "home_service" {
		if input.UserLocation == nil || len(input.UserLocation.Coordinates) < 2 {
			logrus.Warnf("[BookingError] Missing or invalid user coordinates for home service. UserID: %s", userID)
			return nil, errors.New("precise user location coordinates are required for home service")
		}

		if len(carwash.Location.Coordinates) < 2 {
			logrus.Errorf("[BookingError] Carwash %s has invalid location coordinates", input.CarwashID)
			return nil, errors.New("carwash location is improperly configured on the server")
		}

		userLng := input.UserLocation.Coordinates[0]
		userLat := input.UserLocation.Coordinates[1]

		carwashLng := carwash.Location.Coordinates[0]
		carwashLat := carwash.Location.Coordinates[1]

		distance := bs.mapService.GetDistanceKm(userLat, userLng, carwashLat, carwashLng)

		// Fallback for DeliveryRadiusKM if it's 0 (uninitialized)
		effectiveRadius := carwash.DeliveryRadiusKM
		if effectiveRadius <= 0 {
			effectiveRadius = 10 // Match the frontend fallback
		}

		logrus.Infof("[BookingService] Distance check: User(%.6f, %.6f) to Carwash(%.6f, %.6f). Distance: %.2f km. Radius: %d km (Effective: %d)",
			userLat, userLng, carwashLat, carwashLng, distance, carwash.DeliveryRadiusKM, effectiveRadius)

		if float64(distance) > float64(effectiveRadius) {
			return nil, fmt.Errorf("your location is %.1f km away, which is outside our %d km delivery radius", distance, effectiveRadius)
		}
	}

	// Step 6: Save to database
	if err := bs.bookingRepository.CreateBooking(&newBooking); err != nil {
		return nil, err
	}

	// Step 6.5: AUTOMATED ESCROW HOLD
	// Charge the customer now and hold in escrow
	if newBooking.TotalPrice > 0 && !newBooking.IsFree {
		err := bs.walletService.EscrowHold(context.Background(), userID, newBooking.ID.Hex(), newBooking.TotalPrice)
		if err != nil {
			// If hold fails (insufficient funds), we should ideally rollback the booking
			// but for now we log it and return error
			logrus.Errorf("❌ Failed to hold escrow for booking %s: %v", newBooking.ID.Hex(), err)
			return nil, err
		}
		logrus.Infof("💰 Escrow hold of ₦%.2f successful for booking %s", newBooking.TotalPrice, newBooking.ID.Hex())
	}

	// Step 7: Trigger notifications (Both Business Owner & Customer)
	go func() {
		user, err := bs.userRepository.FindUserByID(ownerID)
		if err == nil && bs.notificationService != nil {
			// Notify Business Owner (In-app + Email)
			if !carwash.OwnerID.IsZero() {
				bs.notificationService.SendNewBookingToBusiness(carwash.OwnerID, user.Name, "New Booking")
			} else {
				logrus.Warnf("Carwash %s has no OwnerID, cannot send notification", carwash.Name)
			}

			// Notify Customer (Email)
			bookingTimeStr := newBooking.BookingTime.Format("Jan 2, 2006 at 3:04 PM")
			err = utils.SendBookingConfirmationEmail(user.Email, user.Name, carwash.Name, bookingTimeStr)
			if err != nil {
				logrus.Errorf("Failed to send booking confirmation email to customer: %v", err)
			} else {
				logrus.Infof("Booking confirmation email sent to customer: %s", user.Email)
			}
		}
	}()

	return &newBooking, nil

}

func (bs *BookingService) GetBookingByID(bookingID string) (*models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking ID")
	}

	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

func (bs *BookingService) GetBookingsByUserID(userID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	bookings, err := bs.bookingRepository.GetBookingsByUserID(objID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (bs *BookingService) GetBookingsByCarwashID(carwashID string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	bookings, err := bs.bookingRepository.GetBookingsByCarwashID(objID)
	if err != nil {
		return nil, err
	}

	return bs.enrichBookings(bookings)
}

func (bs *BookingService) UpdateBookingStatus(bookingID string, newStatus string, verificationCode string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// Fetch booking (needed for notifications)
	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return errors.New("booking not found")
	}

	// VALIDATION: Enforce Handshake for Starting the Wash (ONLY for Home Service)
	// We verify the code when moving to "in_progress"
	if newStatus == "in_progress" && booking.BookingType == "home_service" {
		if booking.VerificationCode == "" {
			return errors.New("cannot start: booking has no verification code")
		}
		if booking.VerificationCode != verificationCode {
			return errors.New("invalid verification code. Please request the 4-digit code from the customer to start the wash")
		}
	}

	if err := bs.bookingRepository.UpdateBookingStatus(objID, newStatus); err != nil {
		return err
	}

	// Step 1.5: AUTOMATED SETTLEMENT
	// If the wash is completed, release funds to the carwash wallet
	if newStatus == "completed" && booking.TotalPrice > 0 && !booking.IsFree {
		carwash, err := bs.carWashRepository.GetCarwashByID(booking.CarwashID)
		if err == nil && !carwash.OwnerID.IsZero() {
			err = bs.walletService.EscrowRelease(context.Background(), booking.ID.Hex(), carwash.OwnerID.Hex())
			if err != nil {
				logrus.Errorf("❌ Failed to release escrow for booking %s: %v", booking.ID.Hex(), err)
				// We don't block the status update, but this needs manual reconciliation
			} else {
				logrus.Infof("✅ Automated settlement successful for booking %s. Funds released to owner %s", booking.ID.Hex(), carwash.OwnerID.Hex())
			}
		}
	}

	// Step 2: Generate Verification Code if confirmed and missing
	if newStatus == "confirmed" && booking.VerificationCode == "" {
		code, _ := utils.GenerateNumericCode(4)
		updates := bson.M{"verification_code": code}
		bs.bookingRepository.UpdateBooking(objID, updates)
	}

	// Trigger Notifications (Async)
	go func() {
		if bs.notificationService == nil {
			return
		}

		// Fetch carwash name
		var carwashName string
		cw, err := bs.carWashRepository.GetCarwashByID(booking.CarwashID)
		if err == nil {
			carwashName = cw.Name
		} else {
			carwashName = "The Carwash"
		}

		if newStatus == "confirmed" {
			// In-App + Email (Hybrid Strategy)
			bs.notificationService.SendBookingAccepted(booking, carwashName)
		} else if newStatus == "cancelled" {
			bs.notificationService.SendBookingRejected(booking, "Cancelled by business")
		} else if newStatus == "completed" {
			// Hybrid Strategy: In-App Notification + Detailed Email Receipt
			bs.notificationService.SendPaymentReceipt(booking, carwashName)
		}
	}()

	return nil
}

func (bs *BookingService) CancelBooking(bookingID string) error {
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// 1. Fetch booking to check time
	booking, err := bs.bookingRepository.GetBookingByID(objID)
	if err != nil {
		return errors.New("booking not found")
	}

	// 2. Check if already cancelled or completed
	if booking.Status == "cancelled" {
		return errors.New("booking is already cancelled")
	}
	if booking.Status == "completed" {
		return errors.New("cannot cancel a completed booking")
	}

	// 3. Enforce 24-hour cancellation policy
	// If booking time is within 24 hours from now, deny cancellation
	timeUntilBooking := booking.BookingTime.Sub(time.Now())
	if timeUntilBooking < 24*time.Hour && timeUntilBooking > 0 {
		return errors.New("cancellation not allowed within 24 hours of appointment")
	}

	if err := bs.bookingRepository.CancelBooking(objID); err != nil {
		return err
	}

	// 4. AUTOMATED REFUND
	// Trigger refund of escrowed funds (Fail-soft: we log error but don't block cancellation success)
	if booking.TotalPrice > 0 && !booking.IsFree {
		go func() {
			err := bs.walletService.EscrowRefund(context.Background(), bookingID)
			if err != nil {
				logrus.Errorf("❌ Refund for cancelled booking %s failed: %v", bookingID, err)
			} else {
				logrus.Infof("💰 Automated refund processed for cancelled booking %s", bookingID)
			}
		}()
	}

	return nil
}

func (bs *BookingService) GetBookingsByDate(carwashID string, date time.Time) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	return bs.bookingRepository.GetBookingsByDate(objID, date)
}

func (bs *BookingService) UpdateBooking(userID, bookingID string, updates map[string]interface{}) error {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	bookingObjID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return errors.New("invalid booking ID")
	}

	// Optional: check ownership here if needed

	// Add updatedAt
	updates["updated_at"] = time.Now()

	return bs.bookingRepository.UpdateBooking(bookingObjID, bson.M(updates))
}

// Get Bookings with filter
func (bs *BookingService) GetBookingsByCarwashWithFilters(carwashID string, status string, from, to string) ([]models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(carwashID)
	if err != nil {
		return nil, errors.New("invalid carwash ID")
	}

	var fromDate, toDate time.Time
	if from != "" && to != "" {
		fromDate, err = time.Parse("2006-01-02", from)
		if err != nil {
			return nil, errors.New("invalid from date format")
		}
		toDate, err = time.Parse("2006-01-02", to)
		if err != nil {
			return nil, errors.New("invalid to date format")
		}
		// Adjust toDate to include the full day
		toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, toDate.Location())
	}

	bookings, err := bs.bookingRepository.GetBookingsByCarwashWithFilters(objID, status, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	return bs.enrichBookings(bookings)
}

// Helper to enrich bookings with customer details
func (bs *BookingService) enrichBookings(bookings []models.Booking) ([]models.Booking, error) {
	if len(bookings) == 0 {
		return bookings, nil
	}

	// 1. Collect IDs for Workers and Customers
	userIDs := make([]primitive.ObjectID, 0)
	workerIDs := make([]primitive.ObjectID, 0)
	carwashIDs := make([]primitive.ObjectID, 0)

	userSeen := make(map[primitive.ObjectID]bool)
	workerSeen := make(map[primitive.ObjectID]bool)
	carwashSeen := make(map[primitive.ObjectID]bool)

	for _, b := range bookings {
		if !userSeen[b.UserID] {
			userIDs = append(userIDs, b.UserID)
			userSeen[b.UserID] = true
		}
		if !b.WorkerID.IsZero() && !workerSeen[b.WorkerID] {
			workerIDs = append(workerIDs, b.WorkerID)
			workerSeen[b.WorkerID] = true
		}
		if !carwashSeen[b.CarwashID] {
			carwashIDs = append(carwashIDs, b.CarwashID)
			carwashSeen[b.CarwashID] = true
		}
	}

	// 2. Fetch Users (Customers and Workers)
	allUserIDs := append(userIDs, workerIDs...)
	users, err := bs.userRepository.GetUsersByIDs(allUserIDs)
	if err == nil {
		userMap := make(map[primitive.ObjectID]models.User)
		for _, u := range users {
			userMap[u.ID] = u
		}
		for i := range bookings {
			if user, ok := userMap[bookings[i].UserID]; ok {
				bookings[i].CustomerName = user.Name
				bookings[i].CustomerPhoto = user.ProfilePhoto
			}
			if !bookings[i].WorkerID.IsZero() {
				if worker, ok := userMap[bookings[i].WorkerID]; ok {
					bookings[i].WorkerName = worker.Name
					bookings[i].WorkerPhoto = worker.ProfilePhoto
				}
			}
		}
	}

	// 3. Fetch Carwashes to enrich Services
	carwashes, err := bs.carWashRepository.GetCarwashesByFilter(bson.M{"_id": bson.M{"$in": carwashIDs}})
	if err == nil {
		carwashMap := make(map[primitive.ObjectID]models.Carwash)
		for _, cw := range carwashes {
			carwashMap[cw.ID] = cw
		}

		for i := range bookings {
			if cw, ok := carwashMap[bookings[i].CarwashID]; ok {
				// Map selected service IDs to full service objects
				enrichedServices := make([]models.Service, 0)
				for _, serviceID := range bookings[i].Services {
					for _, s := range cw.Services {
						if s.ID == serviceID {
							enrichedServices = append(enrichedServices, s)
							break
						}
					}
				}
				bookings[i].ServiceDetails = enrichedServices
			}
		}
	}

	return bookings, nil
}

// UpdateWorkerLocation updates the moving coordinates of the service provider
func (bs *BookingService) UpdateWorkerLocation(id string, lat, lng float64) error {
	bookingID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid booking ID format")
	}

	location := models.GeoLocation{
		Type:        "Point",
		Coordinates: []float64{lng, lat}, // MongoDB uses [longitude, latitude]
	}

	updates := bson.M{
		"worker_location": location,
		"updated_at":      time.Now(),
	}

	return bs.bookingRepository.UpdateBooking(bookingID, updates)
}

// CancelExpiredBookings - exposed method to trigger auto-cancellation of pending bookings
func (bs *BookingService) CancelExpiredBookings() (int64, error) {
	return bs.bookingRepository.CancelExpiredBookings()
}

// ReserveSlot attempts to lock a time slot for a user for a short period
func (bs *BookingService) ReserveSlot(carwashID, slotTime, userID string) (bool, error) {
	lockKey := GetLockKey(carwashID, slotTime)
	
	// Attempt to acquire the lock (using config-based 180s)
	success, err := bs.lockService.AcquireLock(lockKey, userID)
	if err != nil {
		logrus.Errorf("❌ Failed to acquire lock for %s: %v", lockKey, err)
		return false, err
	}

	if !success {
		logrus.Warnf("⚠️ Slot %s is already locked by another user", lockKey)
		return false, nil
	}

	logrus.Infof("✅ Slot %s reserved for User %s", lockKey, userID)
	return true, nil
}
