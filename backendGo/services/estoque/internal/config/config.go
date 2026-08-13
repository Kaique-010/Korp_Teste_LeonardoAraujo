package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	RabbitMQURL string
	LogLevel    string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8081"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "15432"),
		DBUser:      getEnv("DB_USER", "estoque"),
		DBPassword:  getEnv("DB_PASSWORD", "estoque"),
		DBName:      getEnv("DB_NAME", "estoque_db"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://korp:korp@localhost:5672/"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
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
