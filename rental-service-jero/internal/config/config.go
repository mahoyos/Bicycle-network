package config

import "os"

type Config struct {
	DatabaseURL   string
	RabbitMQURL   string
	JWTSecret     string
	ServerPort    string
	RunMigrations bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/rentals_db?sslmode=disable"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		RunMigrations: getEnv("RUN_MIGRATIONS", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
