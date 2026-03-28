package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL  string
	RabbitMQURL  string
	JWTPublicKey string
	JWTAlgorithm string
	AppPort      string
	DisableAuth  bool
	RunMigrations bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://user:password@localhost:5433/rental_db"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5673/"),
		JWTPublicKey:  getEnv("JWT_PUBLIC_KEY", ""),
		JWTAlgorithm:  getEnv("JWT_ALGORITHM", "RS256"),
		AppPort:       getEnv("APP_PORT", "8002"),
		DisableAuth:   getBoolEnv("DISABLE_AUTH", false),
		RunMigrations: getBoolEnv("RUN_MIGRATIONS", true),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}
