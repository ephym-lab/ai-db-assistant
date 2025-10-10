package config

import (
	"os"
	"time"
)

type Config struct {
	Port 		  string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiry     time.Duration
	Environment   string
}


func load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mydb?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your_jwt_secret"),
		JWTExpiry:   getEnvAsDuration("JWT_EXPIRY", time.Hour * 24),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}