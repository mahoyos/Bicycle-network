package repository

import (
	"context"

	"github.com/bicycle-network/rental-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BikeRepository interface {
	Upsert(ctx context.Context, bikeID uuid.UUID) error
	Delete(ctx context.Context, bikeID uuid.UUID) error
	Exists(ctx context.Context, bikeID uuid.UUID) (bool, error)
}

type bikeRepository struct {
	db *gorm.DB
}

func NewBikeRepository(db *gorm.DB) BikeRepository {
	return &bikeRepository{db: db}
}

func (r *bikeRepository) Upsert(ctx context.Context, bikeID uuid.UUID) error {
	bike := models.KnownBike{ID: bikeID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&bike).Error
}

func (r *bikeRepository) Delete(ctx context.Context, bikeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.KnownBike{}, "id = ?", bikeID).Error
}

func (r *bikeRepository) Exists(ctx context.Context, bikeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.KnownBike{}).
		Where("id = ?", bikeID).
		Count(&count).Error
	return count > 0, err
}
