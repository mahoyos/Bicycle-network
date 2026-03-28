package models

import (
	"time"

	"github.com/google/uuid"
)

type Rental struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	BicycleID uuid.UUID  `gorm:"type:uuid;not null;index" json:"bicycle_id"`
	Status    string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	StartTime time.Time  `gorm:"not null;default:now()" json:"start_time"`
	EndTime   *time.Time `gorm:"" json:"end_time"`
	Duration  *string    `gorm:"type:interval" json:"duration"`
}

func (Rental) TableName() string {
	return "rentals"
}
