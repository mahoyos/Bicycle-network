package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("RABBITMQ_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("RUN_MIGRATIONS")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://user:password@localhost:5432/rentals_db?sslmode=disable" {
		t.Errorf("unexpected default DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("unexpected default RabbitMQURL: %s", cfg.RabbitMQURL)
	}
	if cfg.ServerPort != "8080" {
		t.Errorf("unexpected default ServerPort: %s", cfg.ServerPort)
	}
	if !cfg.RunMigrations {
		t.Error("expected RunMigrations to default to true")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://custom:pass@db:5432/testdb")
	os.Setenv("JWT_SECRET", "my-secret")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("RUN_MIGRATIONS", "false")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("RUN_MIGRATIONS")
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://custom:pass@db:5432/testdb" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("unexpected ServerPort: %s", cfg.ServerPort)
	}
	if cfg.RunMigrations {
		t.Error("expected RunMigrations to be false")
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	val := getEnv("TEST_KEY", "default")
	if val != "test_value" {
		t.Errorf("expected test_value, got %s", val)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY")

	val := getEnv("NONEXISTENT_KEY", "fallback")
	if val != "fallback" {
		t.Errorf("expected fallback, got %s", val)
	}
}
