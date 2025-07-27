package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	JWTSecret   string
	ServerPort  string
	ServerMode  string
	CORSOrigins []string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "invoice_user"),
		DBPassword:  getEnv("DB_PASSWORD", "invoice_password"),
		DBName:      getEnv("DB_NAME", "invoice_db"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerMode:  getEnv("SERVER_MODE", "debug"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),
	}
}

// getEnv gets environment variable with fallback to default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
