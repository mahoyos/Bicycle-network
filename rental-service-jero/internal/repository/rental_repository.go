package repository

import (
	"context"

	"github.com/bicycle-network/rental-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RentalRepository interface {
	Create(ctx context.Context, rental *models.Rental) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Rental, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Rental, error)
	FindActiveByBicycleID(ctx context.Context, bicycleID uuid.UUID) (*models.Rental, error)
	FindAll(ctx context.Context, limit, offset int) ([]models.Rental, error)
	Update(ctx context.Context, rental *models.Rental) error
}

type rentalRepository struct {
	db *gorm.DB
}

func NewRentalRepository(db *gorm.DB) RentalRepository {
	return &rentalRepository{db: db}
}

func (r *rentalRepository) Create(ctx context.Context, rental *models.Rental) error {
	return r.db.WithContext(ctx).Create(rental).Error
}

func (r *rentalRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rental).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *rentalRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, models.StatusActive).
		First(&rental).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *rentalRepository) FindActiveByBicycleID(ctx context.Context, bicycleID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.WithContext(ctx).
		Where("bicycle_id = ? AND status = ?", bicycleID, models.StatusActive).
		First(&rental).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *rentalRepository) FindAll(ctx context.Context, limit, offset int) ([]models.Rental, error) {
	var rentals []models.Rental
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rentals).Error
	if err != nil {
		return nil, err
	}
	return rentals, nil
}

func (r *rentalRepository) Update(ctx context.Context, rental *models.Rental) error {
	return r.db.WithContext(ctx).Save(rental).Error
}
