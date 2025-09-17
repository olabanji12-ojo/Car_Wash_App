package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"os"
)

//  This struct will store the data we'll inject into context
type AuthContext struct {
	UserID      string
	Email       string
	Role        string
	AccountType string
}

// typed context key (prevents collisions)
type contextKey string

const authKey contextKey = "auth"

//  AuthMiddleware checks for token, validates it, adds user info to context
//  AuthMiddleware checks for token, validates it, adds user info to context
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var tokenString string

        // 1️⃣ Try Authorization header first
        authHeader := r.Header.Get("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
        } else {
            // 2️⃣ If no header, try query param
            tokenString = r.URL.Query().Get("auth_token")
            if tokenString == "" {
                http.Error(w, "Missing auth token", http.StatusUnauthorized)
                return
            }
        }

        // 3️⃣ Validate token
        token, claims, err := utils.ValidateToken(tokenString)
        if err != nil || !token.Valid {
            logrus.Warn("Token invalid or expired: ", err)
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }

        // 4️⃣ Extract claims
        userID, ok1 := claims["user_id"].(string)
        email, ok2 := claims["email"].(string)
        role, ok3 := claims["role"].(string)
        accountType, ok4 := claims["account_type"].(string)

        if !ok1 || !ok2 || !ok3 || !ok4 {
            logrus.Warn("Token claims missing or invalid types")
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        // 5️⃣ Save into context
        authCtx := AuthContext{
            UserID:      userID,
            Email:       email,
            Role:        role,
            AccountType: accountType,
        }
        logrus.WithFields(logrus.Fields{
            "user_id": userID,
            "email":   email,
        }).Info("Authenticated request")

        ctx := context.WithValue(r.Context(), authKey, authCtx)

        // 6️⃣ Continue
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}


func Cors() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{
			
            "https://foam-up.vercel.app",
			"http://localhost:3000", // add this
			
			}, // ✅ exact frontend URL
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Origin"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	})
	return c
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

	secureMiddleware := secure.New(options)

	return secureMiddleware
}
