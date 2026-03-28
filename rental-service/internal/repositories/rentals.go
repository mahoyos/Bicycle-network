package repositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/models"
	"gorm.io/gorm"
)

type RentalsRepository struct {
	db *gorm.DB
}

func NewRentalsRepository(db *gorm.DB) *RentalsRepository {
	return &RentalsRepository{db: db}
}

func (r *RentalsRepository) Create(userID, bicycleID uuid.UUID) (*models.Rental, error) {
	rental := &models.Rental{
		UserID:    userID,
		BicycleID: bicycleID,
		Status:    "active",
		StartTime: time.Now(),
	}
	if err := r.db.Create(rental).Error; err != nil {
		return nil, err
	}
	return rental, nil
}

func (r *RentalsRepository) FindActiveByUserID(userID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.Where("user_id = ? AND status = 'active'", userID).First(&rental).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rental, nil
}

func (r *RentalsRepository) FindActiveByBicycleID(bicycleID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.Where("bicycle_id = ? AND status = 'active'", bicycleID).First(&rental).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rental, nil
}

func (r *RentalsRepository) Finalize(rentalID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.First(&rental, "id = ?", rentalID).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	duration := now.Sub(rental.StartTime)
	durationStr := fmt.Sprintf("%d hours %d minutes %d seconds",
		int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

	rental.Status = "finalized"
	rental.EndTime = &now
	rental.Duration = &durationStr

	if err := r.db.Save(&rental).Error; err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *RentalsRepository) GetByID(rentalID uuid.UUID) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.First(&rental, "id = ?", rentalID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rental, nil
}

// --- Pending Deletes ---

func (r *RentalsRepository) CreatePendingDelete(bicycleID uuid.UUID) error {
	pd := &models.PendingDelete{
		BicycleID: bicycleID,
	}
	// Use FirstOrCreate to avoid duplicates
	return r.db.Where("bicycle_id = ? AND processed = false", bicycleID).
		FirstOrCreate(pd).Error
}

func (r *RentalsRepository) FindPendingDeleteByBicycleID(bicycleID uuid.UUID) (*models.PendingDelete, error) {
	var pd models.PendingDelete
	err := r.db.Where("bicycle_id = ? AND processed = false", bicycleID).First(&pd).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &pd, nil
}

func (r *RentalsRepository) MarkPendingDeleteProcessed(bicycleID uuid.UUID) error {
	return r.db.Model(&models.PendingDelete{}).
		Where("bicycle_id = ? AND processed = false", bicycleID).
		Update("processed", true).Error
}
