package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/database"

	// "github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/routes"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)



func main() {


	err := godotenv.Load()
	if err != nil {
		logrus.Info("No .env file found, using defaults")
	}

	port := os.Getenv("PORT")

	fmt.Println("Hello World")

	database.ConnectDB()
	database.InitCollections()

    router := mux.NewRouter()

	// auth routes
	routes.AuthRoutes(router)

	// user routes
	routes.UserRoutes(router)

	// car routes
    routes.CarRoutes(router)

	// carwash routes
	routes.CarwashRoutes(router)

	// service routes
    routes.ServiceRoutes(router)

	// booking routes
	routes.BookingRoutes(router)

	// payment routes

	routes.PaymentRoutes(router)

	// review routes
	routes.ReviewRoutes(router)

	// order routes
	
	routes.OrderRoutes(router)

	// 4. Start HTTP server
	fmt.Println("🌐 Listening on http://localhost:",port)
	http.ListenAndServe(":8080", router)
	
	
		
}


// NOTE, TO CLEAN UP THE ROUTES BECAUSE OF IT'S REDUNDANCY 





