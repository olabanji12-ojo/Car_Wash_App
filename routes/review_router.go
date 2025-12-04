package routes

import (

	// "net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

type ReviewRouter struct {
	reviewController controllers.ReviewController
}

func NewReviewRouter(reviewController controllers.ReviewController) *ReviewRouter {
	return &ReviewRouter{reviewController: reviewController}
}

func (rr *ReviewRouter) ReviewRoutes(router *mux.Router) {
	review := router.PathPrefix("/api/reviews").Subrouter()

	review.Use(middleware.AuthMiddleware)

	// 🔐 Authenticated routes
	review.HandleFunc("", rr.reviewController.LeaveReviewHandler).Methods("POST")              // tested
	review.HandleFunc("/user", rr.reviewController.GetReviewsByUserHandler).Methods("GET")     // tested
	review.HandleFunc("/{id}/reply", rr.reviewController.ReplyToReviewHandler).Methods("POST") // New reply route

	// 🌐 Public access
	review.HandleFunc("/order/{id}", rr.reviewController.GetReviewByOrderIDHandler).Methods("GET")                // tested
	review.HandleFunc("/business/{id}", rr.reviewController.GetReviewsByBusinessIDHandler).Methods("GET")         // tested
	review.HandleFunc("/carwash/{id}/average", rr.reviewController.GetCarwashAverageRatingHandler).Methods("GET") // tested
}
