package config

import (
	"os"
	"time"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	JWTSecret     string
	JWTExpiration time.Duration
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8083"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "15434"),
		DBName:     getEnv("DB_NAME", "auth_db"),
		DBUser:     getEnv("DB_USER", "auth"),
		DBPassword: getEnv("DB_PASSWORD", "auth"),

		JWTSecret:     getEnv("JWT_SECRET", "korp-teste-super-secret-key-2026"),
		JWTExpiration: getDurationEnv("JWT_EXPIRATION", time.Hour),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
