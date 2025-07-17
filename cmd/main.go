package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/urfave/negroni"
	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/routes"
)

func main() {
	
	err := godotenv.Load()
	if err != nil {
		logrus.Info("No .env file found, using defaults")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("✅ Connecting to database...")
	database.ConnectDB()
	database.InitCollections()

	// 1. Create the main router
	router := mux.NewRouter()

	// 2. Register all routes
	routes.InitRoutes(router) // created a function to handle that 

	// 3. Set up Negroni middleware stack
	
	n := negroni.New()
	secureMiddleware := middleware.Secure()

	// 4. Add security, CORS middleware from your package
	n.Use(negroni.NewRecovery()) // handles panics gracefully
    n.UseHandler(secureMiddleware.Handler(n)) // secure headers
	n.Use(middleware.Cors())          // CORS handling
	n.UseHandler(router)              // finally attach your routes

	// 5. Start server
	fmt.Println("🚀 Listening on http://localhost:" + port)
	http.ListenAndServe(":"+port, n)


}

