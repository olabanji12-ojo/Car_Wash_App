package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/olabanji12-ojo/CarWashApp/utils"
)

// 👇 This struct will store the data we'll inject into context
type AuthContext struct {
	UserID string
	Email  string
	Role   string
}

// 🔐 AuthMiddleware checks for token, validates it, adds user info to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 🔎 1. Get token from "Authorization: Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 🔐 2. Validate token using our utils function
		token, claims, err := utils.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			logrus.Warn("Token invalid or expired:", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// ✅ 3. Extract user data from claims
		userID := claims["user_id"].(string)
		email := claims["email"].(string)
		role := claims["role"].(string)

		// 🧠 4. Save user info into context
		authCtx := AuthContext{
			UserID: userID,
			Email:  email,
			Role:   role,
		}

		ctx := context.WithValue(r.Context(), "auth", authCtx)

		// ⏭️ 5. Call next handler, passing in updated request with user context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}



