package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB(dsn string, runMigrations bool) error {
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if runMigrations {
		if err := runSQLMigrations(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	log.Println("Database connection established")
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func CloseDB() {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func CheckDB() bool {
	if db == nil {
		return false
	}
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

func runSQLMigrations() error {
	// Try multiple possible paths for the migration file
	paths := []string{
		"internal/database/migrations/init.sql",
		"migrations/init.sql",
	}

	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		log.Println("SQL migrations executed successfully")
		return nil
	}

	log.Println("No migration file found, skipping SQL migrations")
	return nil
}
