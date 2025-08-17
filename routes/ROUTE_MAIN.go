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

func InitRoutes(router *mux.Router, db *mongo.Database) {
	AuthRoutes(router, InitAuthService(db))
	UserRoutes(router)
	CarRoutes(router)
	CarwashRoutes(router)
	// ServiceRoutes(router)
	BookingRoutes(router)
	PaymentRoutes(router)
	ReviewRoutes(router)
	OrderRoutes(router)
	WorkerRoutes(router)
	NotificationRoutes(router) // Notification system
}
