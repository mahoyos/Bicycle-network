package service

import (
	"context"
	"errors"
	"time"

	"github.com/bicycle-network/rental-service/internal/models"
	"github.com/bicycle-network/rental-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrBikeNotFound      = errors.New("bicycle not found")
	ErrBikeAlreadyRented = errors.New("bicycle is already rented")
	ErrUserHasActive     = errors.New("user already has an active rental")
	ErrRentalNotFound    = errors.New("rental not found")
	ErrNotOwner          = errors.New("rental does not belong to this user")
	ErrNotActive         = errors.New("rental is not active")
)

type RentalService interface {
	CreateRental(ctx context.Context, userID, bicycleID uuid.UUID) (*models.Rental, error)
	FinalizeRental(ctx context.Context, rentalID, userID uuid.UUID) (*models.Rental, error)
	GetActiveRental(ctx context.Context, userID uuid.UUID) (*models.Rental, error)
	ListAllRentals(ctx context.Context, limit, offset int) ([]models.Rental, error)
	CancelRental(ctx context.Context, rentalID uuid.UUID) (*models.Rental, error)
}

type rentalService struct {
	rentalRepo repository.RentalRepository
	bikeRepo   repository.BikeRepository
}

func NewRentalService(
	rentalRepo repository.RentalRepository,
	bikeRepo repository.BikeRepository,
) RentalService {
	return &rentalService{
		rentalRepo: rentalRepo,
		bikeRepo:   bikeRepo,
	}
}

func (s *rentalService) CreateRental(ctx context.Context, userID, bicycleID uuid.UUID) (*models.Rental, error) {
	exists, err := s.bikeRepo.Exists(ctx, bicycleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrBikeNotFound
	}

	_, err = s.rentalRepo.FindActiveByBicycleID(ctx, bicycleID)
	if err == nil {
		return nil, ErrBikeAlreadyRented
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	_, err = s.rentalRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		return nil, ErrUserHasActive
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rental := &models.Rental{
		UserID:    userID,
		BicycleID: bicycleID,
		Status:    models.StatusActive,
		StartTime: time.Now().UTC(),
	}

	if err := s.rentalRepo.Create(ctx, rental); err != nil {
		return nil, err
	}

	return rental, nil
}

func (s *rentalService) FinalizeRental(ctx context.Context, rentalID, userID uuid.UUID) (*models.Rental, error) {
	rental, err := s.rentalRepo.FindByID(ctx, rentalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRentalNotFound
		}
		return nil, err
	}

	if rental.UserID != userID {
		return nil, ErrNotOwner
	}

	if rental.Status != models.StatusActive {
		return nil, ErrNotActive
	}

	now := time.Now().UTC()
	durationSeconds := int(now.Sub(rental.StartTime).Seconds())

	rental.EndTime = &now
	rental.DurationSeconds = &durationSeconds
	rental.Status = models.StatusFinalized
	rental.UpdatedAt = now

	if err := s.rentalRepo.Update(ctx, rental); err != nil {
		return nil, err
	}

	return rental, nil
}

func (s *rentalService) GetActiveRental(ctx context.Context, userID uuid.UUID) (*models.Rental, error) {
	rental, err := s.rentalRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRentalNotFound
		}
		return nil, err
	}
	return rental, nil
}

func (s *rentalService) ListAllRentals(ctx context.Context, limit, offset int) ([]models.Rental, error) {
	return s.rentalRepo.FindAll(ctx, limit, offset)
}

func (s *rentalService) CancelRental(ctx context.Context, rentalID uuid.UUID) (*models.Rental, error) {
	rental, err := s.rentalRepo.FindByID(ctx, rentalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRentalNotFound
		}
		return nil, err
	}

	if rental.Status != models.StatusActive {
		return nil, ErrNotActive
	}

	now := time.Now().UTC()
	rental.Status = models.StatusCancelled
	rental.EndTime = &now
	rental.UpdatedAt = now

	if err := s.rentalRepo.Update(ctx, rental); err != nil {
		return nil, err
	}

	return rental, nil
}
