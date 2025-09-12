package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/routes"
)

func main() {
	// Initialize config (loads env vars)
	config.Init()
	config.InitGoogleOAuth()

	// Get PORT from env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Show some info
	// fmt.Println("Environment:", config.Cfg.Env)
	fmt.Println("Mongo URL:", os.Getenv("MONGO_URI"))
	fmt.Println("DB Name:", os.Getenv("DB_NAME"))

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db := database.ConnectDB()
	database.InitCollections()

	// Initialize router
	router := mux.NewRouter()
	routes.InitRoutes(router, db)

	// Initialize Cloudinary (if you’re using config.Cfg vars inside)
	config.InitCloudinary()

	// Middleware stack
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(middleware.Cors())
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		secureMiddleware := middleware.Secure()
		if err := secureMiddleware.Process(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if next != nil {
			next(w, r)
		}
	}))
	n.UseHandler(router)

	// Start server
	fmt.Println("🚀 Listening on http://localhost:" + port)
	err := http.ListenAndServe(":"+port, n)
	if err != nil {
		logrus.Fatal("Server failed to start: ", err)
	}
}