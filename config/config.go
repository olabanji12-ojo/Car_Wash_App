package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string
	GoogleClientID      string
	GoogleClientSecret  string
	OAuthRedirectURL    string
	FrontendURL         string
	SessionKey          string
	PlatformFee         float64
	BookingLockTime     int
}

var Cfg Config

func Init() {
	// Choose environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	Cfg.Env = env

	// Load file for development (optional for production)
	envFile := ".env." + env
	if env == "development" {
		// fail quietly if file not present
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("warning: %s not found, will use environment variables", envFile)
		}
	} else {
		// for production, prefer environment variables set by the host;
		// but if you have a .env.production during testing you may load it:
		if _, err := os.Stat(envFile); err == nil {
			if err := godotenv.Load(envFile); err != nil {
				log.Printf("warning: failed to load %s: %v", envFile, err)
			}
		}
	}

	// Required vars — fail fast if missing
	Cfg.GoogleClientID = mustGet("GOOGLE_CLIENT_ID")
	Cfg.GoogleClientSecret = mustGet("GOOGLE_CLIENT_SECRET")
	Cfg.OAuthRedirectURL = mustGet("OAUTH_REDIRECT_URL")
	Cfg.FrontendURL = mustGet("FRONTEND_URL")
	Cfg.SessionKey = mustGet("SESSION_KEY")

	// Platform Fee (Defaults to 0.05 if not set)
	feeStr := os.Getenv("PLATFORM_FEE")
	if feeStr == "" {
		Cfg.PlatformFee = 0.05
	} else {
		// Simple conversion, assuming valid float in env
		fmt.Sscanf(feeStr, "%f", &Cfg.PlatformFee)
	}

	// Booking Lock Time (Defaults to 180s if not set)
	lockTimeStr := os.Getenv("BOOKING_LOCK_TIME")
	if lockTimeStr == "" {
		Cfg.BookingLockTime = 180
	} else {
		fmt.Sscanf(lockTimeStr, "%d", &Cfg.BookingLockTime)
	}
}

func mustGet(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("env var %s is required but not set", key))
	}
	return v
}
