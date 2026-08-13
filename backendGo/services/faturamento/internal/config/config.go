package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	EstoqueBaseURL string
	RabbitMQURL    string
	LogLevel       string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8082"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "15433"),
		DBUser:         getEnv("DB_USER", "faturamento"),
		DBPassword:     getEnv("DB_PASSWORD", "faturamento"),
		DBName:         getEnv("DB_NAME", "faturamento_db"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		EstoqueBaseURL: getEnv("ESTOQUE_BASE_URL", "http://localhost:8081"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://korp:korp@localhost:5672/"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
