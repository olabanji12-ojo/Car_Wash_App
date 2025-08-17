package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/routes"
	"github.com/urfave/negroni"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		logrus.Info("No .env file found, using defaults")
	}

	// Get PORT from env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db := database.ConnectDB()
	database.InitCollections()

	// Initialize router
	router := mux.NewRouter()
	routes.InitRoutes(router, db)

	// Initialize Cloudinary
	config.InitCloudinary()

	// Negroni middleware stack
	n := negroni.New()

	n.Use(negroni.NewRecovery()) // Handles panic recovery
	n.Use(middleware.Cors())     // Enable CORS
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		secureMiddleware := middleware.Secure()
		if err := secureMiddleware.Process(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if next != nil {
			next(w, r)
		}
	})) // Custom security headers (not circular)
	n.UseHandler(router) // Final handler: the actual router

	// Start server
	fmt.Println("🚀 Listening on http://localhost:" + port)
	err = http.ListenAndServe(":"+port, n)

	if err != nil {
		logrus.Fatal("Server failed to start: ", err)
	}

}
