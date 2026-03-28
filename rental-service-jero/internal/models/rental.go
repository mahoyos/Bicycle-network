package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusActive    = "active"
	StatusFinalized = "finalized"
	StatusCancelled = "cancelled"
)

type Rental struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	BicycleID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"bicycle_id"`
	Status          string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	StartTime       time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"start_time"`
	EndTime         *time.Time `gorm:"type:timestamptz" json:"end_time"`
	DurationSeconds *int       `gorm:"type:integer" json:"duration_seconds"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

func (r *Rental) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = StatusActive
	}
	if r.StartTime.IsZero() {
		r.StartTime = time.Now().UTC()
	}
	return nil
}
