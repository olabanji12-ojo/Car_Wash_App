package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitAuthService(db *mongo.Database) *controllers.AuthController {
	authService := services.NewAuthService(*repositories.NewUserRepository(db))
	return &controllers.AuthController{AuthService: authService}
}

func InitCarService(db *mongo.Database) *controllers.CarController {
	carService := services.NewCarService(*repositories.NewCarRepository(db))
	return &controllers.CarController{CarService: carService}
}

func InitCarWashService(db *mongo.Database) *controllers.CarWashController {
	carwashService := services.NewCarWashService(*repositories.NewCarWashRepository(db))
	return &controllers.CarWashController{CarWashService: carwashService}
}

func InitBookingService(db *mongo.Database) *controllers.BookingController {
	bookingService := services.NewBookingService(*repositories.NewBookingRepository(db))
	return &controllers.BookingController{BookingService: bookingService}
}

func InitOrderService(db *mongo.Database) *controllers.OrderController {
	orderService := services.NewOrderService(*repositories.NewOrderRepository(db))
	return &controllers.OrderController{OrderService: orderService}
}
func InitReviewService(db *mongo.Database) *controllers.ReviewController {
	reviewService := services.NewReviewService(*repositories.NewReviewRepository(db))
	return controllers.NewReviewController(reviewService)
}

func InitRoutes(router *mux.Router, db *mongo.Database) {
	AuthRoutes(router, InitAuthService(db))
	UserRoutes(router)

	// Initialize CarRouter and set up car routes
	carController := InitCarService(db)
	carRouter := NewCarRouter(*carController)
	carRouter.CarRoutes(router)

	// Initialize CarWashRouter and set up car routes

	carwashController := InitCarWashService(db)
	carwashRouter := NewCarWashRouter(*carwashController)
	carwashRouter.CarwashRoutes(router)

	// Initialize BookingRouter and set up booking routes
	bookingController := InitBookingService(db)
	bookingRouter := NewBookingRouter(*bookingController)
	bookingRouter.BookingRoutes(router)
	
	// Initialize OrderRouter and set up order routes
	orderController := InitOrderService(db)
	OrderRouter := NewOrderRouter(orderController)
	OrderRouter.OrderRoutes(router)

    // Initialize ReviewRouter and set up order routes
    
	reviewController := InitReviewService(db)
	ReviewRouter := NewReviewRouter(*reviewController)
	ReviewRouter.ReviewRoutes(router)
    	
	
	PaymentRoutes(router)
	
	WorkerRoutes(router)
	NotificationRoutes(router) // Notification system
    
}

