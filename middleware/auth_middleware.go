package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/secure"

	"os"
)

// This struct will store the data we'll inject into context
type AuthContext struct {
	UserID      string
	Email       string
	Role        string
	AccountType string
}

// typed context key (prevents collisions)
type contextKey string

const authKey contextKey = "auth"

// ✅ Helper function that controllers can use for safe context retrieval
func GetAuthContext(r *http.Request) (*AuthContext, bool) {
	authValue := r.Context().Value(authKey)
	if authValue == nil {
		logrus.Warn("Auth context not found in request")
		return nil, false
	}

	authCtx, ok := authValue.(AuthContext)
	if !ok {
		logrus.Warn("Auth context type assertion failed")
		return nil, false
	}

	return &authCtx, true
}

// ✅ Alternative helper that matches your current controller pattern exactly
func GetAuthContextDirect(r *http.Request) (AuthContext, error) {
	authValue := r.Context().Value("auth")
	if authValue == nil {
		return AuthContext{}, fmt.Errorf("authentication context not found")
	}

	authCtx, ok := authValue.(AuthContext)
	if !ok {
		return AuthContext{}, fmt.Errorf("invalid authentication context type")
	}

	return authCtx, nil
}

// AuthMiddleware checks for token, validates it, adds user info to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Try cookie
			cookie, err := r.Cookie("jwt")
			if err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			} else {
				// Try query param
				tokenString = r.URL.Query().Get("auth_token")
			}
		}

		if tokenString == "" {
			logrus.Warn("No auth token provided")
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		token, claims, err := utils.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			logrus.WithError(err).Warn("Token invalid or expired")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		userID, ok1 := claims["user_id"].(string)
		email, ok2 := claims["email"].(string)
		role, ok3 := claims["role"].(string)
		accountType, ok4 := claims["account_type"].(string)

		if !ok1 || !ok2 || !ok3 || !ok4 || userID == "" || email == "" || role == "" || accountType == "" {
			logrus.Warn("Invalid or missing token claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		authCtx := AuthContext{
			UserID:      userID,
			Email:       email,
			Role:        role,
			AccountType: accountType,
		}

		ctx := context.WithValue(r.Context(), authKey, authCtx)
		ctx = context.WithValue(ctx, "auth", authCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS and Secure middleware unchanged
func Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
			"https://car-wash-frontend-ten.vercel.app",
		},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Origin", "X-CSRF-Token"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		Debug:            os.Getenv("ENVIRONMENT") != "production",
	})
}

func Secure() *secure.Secure {
	options := secure.Options{
		BrowserXssFilter:     true,
		ContentTypeNosniff:   true,
		FrameDeny:            false,
		SSLForceHost:         false,
		STSIncludeSubdomains: true,
		STSPreload:           true,
	}
	if os.Getenv("ENVIRONMENT") != "production" {
		options.IsDevelopment = true
	}
	return secure.New(options)
}
