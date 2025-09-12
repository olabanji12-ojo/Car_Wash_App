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
}

func mustGet(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("env var %s is required but not set", key))
	}
	return v
}
