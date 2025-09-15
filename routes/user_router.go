package routes

import (
	

	"github.com/gorilla/mux"

	"github.com/olabanji12-ojo/CarWashApp/controllers"

	// "github.com/olabanji12-ojo/CarWashApp/middleware"

)


func UserRoutes(router *mux.Router) {
	
	
	//  User Routes

	userRouter := router.PathPrefix("/api/user").Subrouter()

	// userRouter.Use(middleware.AuthMiddleware)

	userRouter.HandleFunc("/{id}", controllers.GetUserProfile).Methods("GET") // tested

	userRouter.HandleFunc("/me/", controllers.GetCurrentUser).Methods("GET")
    
	userRouter.HandleFunc("/{id}", controllers.UpdateUserProfile).Methods("PUT") // tested

	userRouter.HandleFunc("/{id}", controllers.DeleteUser).Methods("DELETE") // tested
    
	userRouter.HandleFunc("/{id}/role", controllers.GetUserRole).Methods("GET")  // tested

	userRouter.HandleFunc("/{id}/loyalty", controllers.GetLoyaltyPoints).Methods("GET") // tested

	userRouter.HandleFunc("/{id}/public", controllers.GetPublicUser).Methods("GET") // tested

	//  Business Route for Workers
	businessRouter := router.PathPrefix("/api/business").Subrouter()
	businessRouter.HandleFunc("/{id}/workers", controllers.GetWorkersForBusiness).Methods("GET") // pending
    
	
}
