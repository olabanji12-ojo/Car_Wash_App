package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	// "github.com/olabanji12-ojo/CarWashApp/middleware"
)

// UserRouter handles user-related routing
type UserRouter struct {
	userController *controllers.UserController
}

// NewUserRouter creates a new UserRouter instance
func NewUserRouter(userController *controllers.UserController) *UserRouter {
	return &UserRouter{userController: userController}
}

// UserRoutes sets up all user-related routes
func (ur *UserRouter) UserRoutes(router *mux.Router) {
	
	// User Routes
	userRouter := router.PathPrefix("/api/user").Subrouter()

	// Apply auth middleware to all user routes if needed
	// userRouter.Use(middleware.AuthMiddleware)

	// Basic user operations
	userRouter.HandleFunc("/{id}", ur.userController.GetUserProfile).Methods("GET")                 // tested
	userRouter.HandleFunc("/callback/me", ur.userController.GetCurrentUser).Methods("GET")          // tested
	userRouter.HandleFunc("/{id}", ur.userController.UpdateUserProfile).Methods("PUT")              // tested
	userRouter.HandleFunc("/{id}", ur.userController.DeleteUser).Methods("DELETE")                  // tested
	userRouter.HandleFunc("/{id}/role", ur.userController.GetUserRole).Methods("GET")               // tested
	userRouter.HandleFunc("/{id}/loyalty", ur.userController.GetLoyaltyPoints).Methods("GET")       // tested
	userRouter.HandleFunc("/{id}/public", ur.userController.GetPublicUser).Methods("GET")           // tested

	// Address management routes
	userRouter.HandleFunc("/{id}/addresses", ur.userController.GetUserAddresses).Methods("GET")                       // new
	userRouter.HandleFunc("/{id}/addresses", ur.userController.AddUserAddress).Methods("POST")                        // MVP
	userRouter.HandleFunc("/{id}/addresses/{address_id}", ur.userController.UpdateUserAddress).Methods("PUT")         // new
	userRouter.HandleFunc("/{id}/addresses/{address_id}", ur.userController.DeleteUserAddress).Methods("DELETE")      // MVP
	userRouter.HandleFunc("/{id}/addresses/{address_id}/default", ur.userController.SetDefaultAddress).Methods("PATCH") // new

	// Profile photo routes
	userRouter.HandleFunc("/{id}/photo", ur.userController.UploadProfilePhoto).Methods("POST")      // MVP
	userRouter.HandleFunc("/{id}/photo", ur.userController.DeleteProfilePhoto).Methods("DELETE")    // MVP

	//  Business Route for Workers
	// businessRouter := router.PathPrefix("/api/business").Subrouter()
	// businessRouter.HandleFunc("/{id}/workers", ur.userController.GetWorkersForBusiness).Methods("GET") // pending
    
	
}
	
	

