package main

import (
    "fmt"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "github.com/gorilla/csrf"
    "github.com/sirupsen/logrus"
    "github.com/urfave/negroni"
    "github.com/olabanji12-ojo/CarWashApp/config"
    "github.com/olabanji12-ojo/CarWashApp/database"
    "github.com/olabanji12-ojo/CarWashApp/middleware"
    "github.com/olabanji12-ojo/CarWashApp/routes"
    "github.com/olabanji12-ojo/CarWashApp/services/geocoding/google"
    "github.com/joho/godotenv"
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
    
    router := mux.NewRouter()
    routes.InitRoutes(router, db, geocoder) // Pass geocoder to routes
    config.InitCloudinary()

    // CSRF Protection - gorilla/csrf compatible with mux
    csrfSecret := []byte(os.Getenv("CSRF_SECRET"))
    if len(csrfSecret) == 0 {
        csrfSecret = []byte("default-secret-change-in-production")
        logrus.Warn("⚠️ Using default CSRF secret. Set CSRF_SECRET in .env for production!")
    }
   
    csrfMiddleware := csrf.Protect(
        csrfSecret,
        csrf.Secure(false), // Set to true in production with HTTPS
        csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte(`{"error": "CSRF token mismatch"}`))
        })),
    )

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
   
    // Create a new router for non-API routes that need CSRF protection
    nonAPIRouter := mux.NewRouter()
   
    // Apply CSRF protection only to non-API routes
    csrfProtected := csrfMiddleware(nonAPIRouter)
   
    // Create a new root router
    rootRouter := mux.NewRouter()
   
    // Mount the API router (no CSRF)
    rootRouter.PathPrefix("/api").Handler(router)
   
    // Mount the non-API router (with CSRF)
    rootRouter.PathPrefix("/").Handler(csrfProtected)
   
    // Use the root router
    n.UseHandler(rootRouter)

    fmt.Println("🚀 Listening on http://localhost:" + port)
    if err := http.ListenAndServe(":"+port, n); err != nil {
        logrus.Fatal("Server failed to start: ", err)
    }
}