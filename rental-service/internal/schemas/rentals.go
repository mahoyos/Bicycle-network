package schemas

import (
	"time"

	"github.com/google/uuid"
)

type CreateRentalRequest struct {
	BicycleID uuid.UUID `json:"bicycle_id" binding:"required"`
}

type RentalResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	BicycleID uuid.UUID  `json:"bicycle_id"`
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Duration  *string    `json:"duration,omitempty"`
}

type ActiveRentalResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	BicycleID       uuid.UUID  `json:"bicycle_id"`
	Status          string     `json:"status"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	Duration        *string    `json:"duration,omitempty"`
	DurationSoFar   string     `json:"duration_so_far"`
}
