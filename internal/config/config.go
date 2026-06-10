package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                     string
    AppPort                    string
	DatabasePath               string
	FirebaseServiceAccountPath string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: .env file not found, failed to read environment variable")
	}

	return &Config {
		AppEnv: getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/fairshare.db"),
		FirebaseServiceAccountPath: getEnv("FIREBASE_SERVICE_ACCOUNT_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
