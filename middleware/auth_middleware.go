package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	"github.com/rs/cors"
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

        // 1️⃣ Try Authorization header first
        authHeader := r.Header.Get("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
        } else {
            // 2️⃣ If no header, try query param
            tokenString = r.URL.Query().Get("auth_token")
            if tokenString == "" {
                logrus.Warn("No auth token provided")
                http.Error(w, "Missing auth token", http.StatusUnauthorized)
                return
            }
        }

        // 3️⃣ Validate token
        token, claims, err := utils.ValidateToken(tokenString)
        logrus.WithFields(logrus.Fields{
            "token_valid": token != nil && token.Valid,
            "claims_count": len(claims),
        }).Info("Token validation attempt")

        if err != nil || !token.Valid {
            logrus.WithError(err).Warn("Token invalid or expired")
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }

        // 4️⃣ Extract claims with detailed logging
        userID, ok1 := claims["user_id"].(string)
        email, ok2 := claims["email"].(string)
        role, ok3 := claims["role"].(string)
        accountType, ok4 := claims["account_type"].(string)

        // Log detailed claim extraction info
        logrus.WithFields(logrus.Fields{
            "user_id_ok": ok1,
            "email_ok": ok2,
            "role_ok": ok3,
            "account_type_ok": ok4,
            "user_id_val": userID,
            "email_val": email,
            "role_val": role,
            "account_type_val": accountType,
        }).Info("Claims extraction")

        if !ok1 || !ok2 || !ok3 || !ok4 {
            logrus.WithFields(logrus.Fields{
                "user_id_type": fmt.Sprintf("%T", claims["user_id"]),
                "email_type": fmt.Sprintf("%T", claims["email"]),
                "role_type": fmt.Sprintf("%T", claims["role"]),
                "account_type_type": fmt.Sprintf("%T", claims["account_type"]),
            }).Warn("Token claims missing or invalid types")
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        // 5️⃣ Create and validate AuthContext
        authCtx := AuthContext{
            UserID:      userID,
            Email:       email,
            Role:        role,
            AccountType: accountType,
        }

        // Validate that all fields are non-empty
        if authCtx.UserID == "" || authCtx.Email == "" || authCtx.Role == "" || authCtx.AccountType == "" {
            logrus.WithFields(logrus.Fields{
                "user_id_empty": authCtx.UserID == "",
                "email_empty": authCtx.Email == "",
                "role_empty": authCtx.Role == "",
                "account_type_empty": authCtx.AccountType == "",
            }).Warn("Auth context has empty required fields")
            http.Error(w, "Invalid token claims - empty fields", http.StatusUnauthorized)
            return
        }
        
        logrus.WithFields(logrus.Fields{
            "user_id": userID,
            "email":   email,
            "role":    role,
            "account_type": accountType,
        }).Info("Successfully authenticated request")

        // 6️⃣ Set context with both keys for compatibility
        ctx := context.WithValue(r.Context(), authKey, authCtx)
        ctx = context.WithValue(ctx, "auth", authCtx) // Also set with string key for compatibility

        // 7️⃣ Verify context was set correctly
        testValue := ctx.Value("auth")
        if testValue == nil {
            logrus.Error("Failed to set auth context in request!")
            http.Error(w, "Internal authentication error", http.StatusInternalServerError)
            return
        }

        // Verify type assertion will work
        if _, ok := testValue.(AuthContext); !ok {
            logrus.Error("Auth context type assertion will fail!")
            http.Error(w, "Internal authentication error", http.StatusInternalServerError)
            return
        }

        logrus.Info("Auth context successfully set and verified")

        // 8️⃣ Continue to next handler
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
