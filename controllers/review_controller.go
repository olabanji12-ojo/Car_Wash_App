package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewController struct {
	reviewService *services.ReviewService
}

func NewReviewController(reviewService *services.ReviewService) *ReviewController {
	return &ReviewController{reviewService: reviewService}
}

// 1. Leave a review
func (rc *ReviewController) LeaveReviewHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	// Use a separate input struct to handle string IDs from JSON
	var input struct {
		CarwashID   string   `json:"carwash_id"`
		OrderID     string   `json:"order_id,omitempty"`
		Rating      int      `json:"rating"`
		Accuracy    int      `json:"accuracy"`
		Cleanliness int      `json:"cleanliness"`
		Comment     string   `json:"comment,omitempty"`
		Photos      []string `json:"photos,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid review data")
		return
	}

	// Convert carwash_id to ObjectID
	carwashObjID, err := primitive.ObjectIDFromHex(input.CarwashID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carwash ID")
		return
	}

	// Convert order_id if provided
	var orderObjID *primitive.ObjectID
	if input.OrderID != "" {
		oid, err := primitive.ObjectIDFromHex(input.OrderID)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid order ID")
			return
		}
		orderObjID = &oid
	}

	// Create review model
	reviewModel := models.Review{
		CarwashID:   carwashObjID,
		OrderID:     orderObjID,
		Rating:      input.Rating,
		Accuracy:    input.Accuracy,
		Cleanliness: input.Cleanliness,
		Comment:     input.Comment,
		Photos:      input.Photos,
	}

	if err := reviewModel.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	review, err := rc.reviewService.CreateReview(userID, reviewModel)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, review)
}

// 2. Get review by order ID
func (rc *ReviewController) GetReviewByOrderIDHandler(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["id"]

	review, err := rc.reviewService.GetReviewByOrderID(orderID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, review)
}

// 3. Get reviews by current user
func (rc *ReviewController) GetReviewsByUserHandler(w http.ResponseWriter, r *http.Request) {
	authCtx := r.Context().Value("auth").(middleware.AuthContext)
	userID := authCtx.UserID

	reviews, err := rc.reviewService.GetReviewsByUserID(userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, reviews)
}

// 4. Get reviews by carwash/business ID
func (rc *ReviewController) GetReviewsByBusinessIDHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["id"]

	reviews, err := rc.reviewService.GetReviewsByCarwashID(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, reviews)
}

// 5. Get average rating for carwash
func (rc *ReviewController) GetCarwashAverageRatingHandler(w http.ResponseWriter, r *http.Request) {
	carwashID := mux.Vars(r)["id"]

	avg, err := rc.reviewService.GetAverageRating(carwashID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"carwash_id": carwashID,
		"average":    avg,
	})
}

// 6. Reply to a review
func (rc *ReviewController) ReplyToReviewHandler(w http.ResponseWriter, r *http.Request) {
	reviewID := mux.Vars(r)["id"]

	var input struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := rc.reviewService.ReplyToReview(reviewID, input.Response); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Response added successfully"})
}
