package routes

import (
	"github.com/gorilla/mux"
)

func InitRoutes(router *mux.Router) {
	AuthRoutes(router)
	UserRoutes(router)
	CarRoutes(router)
	CarwashRoutes(router)
	ServiceRoutes(router)
	BookingRoutes(router)
	PaymentRoutes(router)
	ReviewRoutes(router)
	OrderRoutes(router)
	WorkerRoutes(router)
	NotificationRoutes(router) // Notification system
}



