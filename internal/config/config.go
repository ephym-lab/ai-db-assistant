package config

import (
	"os"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiry     time.Duration
	Environment   string
	ProxyServerURL string
}


func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://neondb_owner:npg_GQAxtdw0BK2i@ep-round-recipe-aexn2xjv-pooler.c-2.us-east-2.aws.neon.tech/neondb?sslmode=require&channel_binding=require"),
		JWTSecret:      getEnv("JWT_SECRET", "your_jwt_secret"),
		JWTExpiry:      time.Hour * 24, // 24 hours
		Environment:    getEnv("ENVIRONMENT", "development"),
		ProxyServerURL: getEnv("PROXY_SERVER_URL", "http://localhost:8000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}