package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"strings"
	"net/http" 
	// "errors" 

	// "fmt"
)

// 🔐 Snippet for generating a secure random JWT secret (for dev use)
/*
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSecureSecret() {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(b))
}
*/

var jwtSecret []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.Warn(".env file not found, reading JWT secret from system")
	}

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		logrus.Fatal("JWT_SECRET not set in environment variables")
	}
}

//  GenerateToken creates a JWT for a given user with a 24-hour expiration
func GenerateToken(userID, email, role string, accountType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"account_type": accountType,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // token expires in 24h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		logrus.Error("Error signing token: ", err)
		return "", err
	}

	return signedToken, nil
}

//  ValidateToken parses and validates a JWT and returns token + claims
func ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure it's signed with the right method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logrus.Error("Unexpected signing method")
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		logrus.Warn("Token parse error: ", err)
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logrus.Warn("Invalid token claims")
		return nil, nil, errors.New("invalid token claims")
	}

	return token, claims, nil
}

//  ExtractClaims extracts claims only (shortcut for controllers if needed)
func ExtractClaims(tokenString string) (map[string]interface{}, error) {
	_, claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil

}


func GetUserIDFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	// Expect "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	tokenStr := parts[1]

	// Validate & extract claims
	claims, err := ExtractClaims(tokenStr)
	if err != nil {
		return "", err
	}

	// Get user_id claim
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id not found in token claims")
	}

	return userID, nil
}

