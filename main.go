package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
	"github.com/olabanji12-ojo/CarWashApp/routes"
	"github.com/olabanji12-ojo/CarWashApp/services/geocoding/google"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Println("No .env file found")
	}

	config.InitGoogleOAuth()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Mongo URL:", os.Getenv("MONGO_URI"))
	fmt.Println("DB Name:", os.Getenv("DB_NAME"))
	fmt.Println("🔌 Connecting to database...")

	db := database.ConnectDB()
	database.InitCollections()

	// Initialize geocoder
	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		logrus.Fatal("❌ GOOGLE_MAPS_API_KEY environment variable is not set")
	}
	geocoder := google.NewGoogleMapsGeocoder(googleMapsAPIKey)
	logrus.Println("✅ Google Maps Geocoder initialized")

	// Create a single main router
	mainRouter := mux.NewRouter()
	routes.InitRoutes(mainRouter, db, geocoder) // Pass geocoder to routes
	config.InitCloudinary()

	csrfSecret := []byte(os.Getenv("CSRF_SECRET"))
	if len(csrfSecret) == 0 {
		csrfSecret = []byte("default-secret-change-in-production")
		logrus.Warn("⚠️ Using default CSRF secret. Set CSRF_SECRET in .env for production!")
	}

	csrfMiddleware := csrf.Protect(
		csrfSecret,
		csrf.Secure(os.Getenv("ENVIRONMENT") == "production"), // Use env var
		csrf.Path("/"), // Apply to all routes, API routes will ignore it
	)

	// --- Middleware Chain with Negroni ---
	n := negroni.New()
	// 1. Recovery middleware (first)
	n.Use(negroni.NewRecovery())
	// 2. CORS middleware (runs before auth)
	n.Use(middleware.Cors())
	// 3. Secure middleware
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		secureMiddleware := middleware.Secure()
		secureMiddleware.HandlerFuncWithNext(w, r, next)
	}))

	// We will use a custom handler to decide when to apply CSRF
	// API routes (stateless, token-based) do not need CSRF.
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for API routes (they use JWT tokens)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			mainRouter.ServeHTTP(w, r)
			return
		}
		// Apply CSRF to non-API routes
		csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mainRouter.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
	n.UseHandler(finalHandler)

	fmt.Println("🚀 Listening on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, n); err != nil {
		logrus.Fatal("Server failed to start: ", err)
	}
}
