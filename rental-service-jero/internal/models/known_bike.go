package models

import (
	"time"

	"github.com/google/uuid"
)

type KnownBike struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}
