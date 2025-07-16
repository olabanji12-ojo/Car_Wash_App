package routes

import (

    // "net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

func ReviewRoutes(router *mux.Router) {
	review := router.PathPrefix("/api/reviews").Subrouter()

	review.Use(middleware.AuthMiddleware)

	// 🔐 Authenticated routes
	review.HandleFunc("", controllers.LeaveReviewHandler).Methods("POST") // tested
	review.HandleFunc("/user", controllers.GetReviewsByUserHandler).Methods("GET") // tested

	// 🌐 Public access
	review.HandleFunc("/order/{id}", controllers.GetReviewByOrderIDHandler).Methods("GET") // tested
	review.HandleFunc("/business/{id}", controllers.GetReviewsByBusinessIDHandler).Methods("GET") // tested
	review.HandleFunc("/carwash/{id}/average", controllers.GetCarwashAverageRatingHandler).Methods("GET") // tested
}
