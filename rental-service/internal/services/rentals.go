package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/models"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/schemas"
)

// Repository interface for dependency injection and testing
type RentalRepository interface {
	Create(userID, bicycleID uuid.UUID) (*models.Rental, error)
	FindActiveByUserID(userID uuid.UUID) (*models.Rental, error)
	FindActiveByBicycleID(bicycleID uuid.UUID) (*models.Rental, error)
	Finalize(rentalID uuid.UUID) (*models.Rental, error)
	GetByID(rentalID uuid.UUID) (*models.Rental, error)
}

// Messaging interface for dependency injection and testing
type EventPublisher interface {
	PublishBicycleReturned(bicycleID string) error
}

type RentalsService struct {
	repo      RentalRepository
	messaging EventPublisher
}

func NewRentalsService(repo RentalRepository, mq EventPublisher) *RentalsService {
	return &RentalsService{repo: repo, messaging: mq}
}

func (s *RentalsService) CreateRental(userID uuid.UUID, req schemas.CreateRentalRequest) (*models.Rental, error) {
	// FR-27: Check user has no active rental
	existing, err := s.repo.FindActiveByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	if existing != nil {
		return nil, &ConflictError{Message: "User already has an active rental"}
	}

	// FR-26: Validate bicycle availability internally
	rented, err := s.repo.FindActiveByBicycleID(req.BicycleID)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	if rented != nil {
		return nil, &ConflictError{Message: "Bicycle is not available"}
	}

	// FR-23: Create the rental
	rental, err := s.repo.Create(userID, req.BicycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create rental: %w", err)
	}
	return rental, nil
}

func (s *RentalsService) FinalizeRental(userID uuid.UUID, rentalID uuid.UUID) (*models.Rental, error) {
	rental, err := s.repo.GetByID(rentalID)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	if rental == nil {
		return nil, &NotFoundError{Message: "Rental not found"}
	}

	if rental.UserID != userID {
		return nil, &ForbiddenError{Message: "Rental does not belong to user"}
	}

	if rental.Status != "active" {
		return nil, &ConflictError{Message: "Rental is not active"}
	}

	finalized, err := s.repo.Finalize(rentalID)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize rental: %w", err)
	}

	// FR-30: Publish RETURNED event
	if s.messaging != nil {
		_ = s.messaging.PublishBicycleReturned(rental.BicycleID.String())
	}

	return finalized, nil
}

func (s *RentalsService) GetActiveRental(userID uuid.UUID) (*schemas.ActiveRentalResponse, error) {
	rental, err := s.repo.FindActiveByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	if rental == nil {
		return nil, &NotFoundError{Message: "No active rental found"}
	}

	elapsed := time.Since(rental.StartTime)
	durationSoFar := fmt.Sprintf("%d hours %d minutes %d seconds",
		int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)

	return &schemas.ActiveRentalResponse{
		ID:            rental.ID,
		UserID:        rental.UserID,
		BicycleID:     rental.BicycleID,
		Status:        rental.Status,
		StartTime:     rental.StartTime,
		EndTime:       rental.EndTime,
		Duration:      rental.Duration,
		DurationSoFar: durationSoFar,
	}, nil
}

// Custom error types

type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

type NotFoundError struct{ Message string }

func (e *NotFoundError) Error() string { return e.Message }

type ForbiddenError struct{ Message string }

func (e *ForbiddenError) Error() string { return e.Message }
