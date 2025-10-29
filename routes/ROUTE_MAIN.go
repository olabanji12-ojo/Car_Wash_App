package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/services/geocoding"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitAuthService(db *mongo.Database) *controllers.AuthController {
	authService := services.NewAuthService(*repositories.NewUserRepository(db))
	return &controllers.AuthController{AuthService: authService}
}

func InitUserService(db *mongo.Database) *controllers.UserController {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	return controllers.NewUserController(userService)
}

func InitWorkerService(db *mongo.Database) *controllers.WorkerController {
	userRepo := repositories.NewUserRepository(db)
	workerRepo := repositories.NewWorkerRepository(db)
	workerService := services.NewWorkerService(userRepo, workerRepo)
	userService := services.NewUserService(userRepo)
	return controllers.NewWorkerController(workerService, userService)
}

func InitCarService(db *mongo.Database) *controllers.CarController {
	carService := services.NewCarService(*repositories.NewCarRepository(db))
	return &controllers.CarController{CarService: carService}
}

func InitCarWashService(db *mongo.Database, geocoder geocoding.Geocoder) *controllers.CarWashController {
	carwashRepo := repositories.NewCarWashRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)
	carwashService := services.NewCarWashService(*carwashRepo, *bookingRepo, geocoder)
	
	// Also initialize UserService for UpdateUserCarwashID
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	
	return controllers.NewCarWashController(carwashService, userService)
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

func InitRoutes(router *mux.Router, db *mongo.Database, geocoder geocoding.Geocoder) {
	AuthRoutes(router, InitAuthService(db))

	// Initialize UserRouter and set up user routes
	userController := InitUserService(db)
	userRouter := NewUserRouter(userController)
	userRouter.UserRoutes(router)

	// Initialize CarRouter and set up car routes
	carController := InitCarService(db)
	carRouter := NewCarRouter(*carController)
	carRouter.CarRoutes(router)

	// Initialize CarWashRouter and set up car wash routes (now with geocoder)
	carwashController := InitCarWashService(db, geocoder)
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

	// Initialize ReviewRouter and set up review routes
	reviewController := InitReviewService(db)
	ReviewRouter := NewReviewRouter(*reviewController)
	ReviewRouter.ReviewRoutes(router)

	// Initialize WorkerRouter and set up worker routes
	workerController := InitWorkerService(db)
	workerRouter := NewWorkerRouter(workerController)
	workerRouter.WorkerRoutes(router)
	
	PaymentRoutes(router)
	NotificationRoutes(router) // Notification system
}