package service

import (
	"context"
	"testing"
	"time"

	"github.com/bicycle-network/rental-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Mock BikeRepository ---

type mockBikeRepo struct {
	existsResult bool
	existsErr    error
}

func (m *mockBikeRepo) Upsert(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockBikeRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockBikeRepo) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.existsResult, m.existsErr
}

// --- Mock RentalRepository ---

type mockRentalRepo struct {
	createErr           error
	findByIDResult      *models.Rental
	findByIDErr         error
	activeByUserResult  *models.Rental
	activeByUserErr     error
	activeByCycleResult *models.Rental
	activeByCycleErr    error
	findAllResult       []models.Rental
	findAllErr          error
	updateErr           error
}

func (m *mockRentalRepo) Create(_ context.Context, rental *models.Rental) error {
	if m.createErr != nil {
		return m.createErr
	}
	if rental.ID == uuid.Nil {
		rental.ID = uuid.New()
	}
	return nil
}

func (m *mockRentalRepo) FindByID(_ context.Context, _ uuid.UUID) (*models.Rental, error) {
	return m.findByIDResult, m.findByIDErr
}

func (m *mockRentalRepo) FindActiveByUserID(_ context.Context, _ uuid.UUID) (*models.Rental, error) {
	return m.activeByUserResult, m.activeByUserErr
}

func (m *mockRentalRepo) FindActiveByBicycleID(_ context.Context, _ uuid.UUID) (*models.Rental, error) {
	return m.activeByCycleResult, m.activeByCycleErr
}

func (m *mockRentalRepo) FindAll(_ context.Context, _, _ int) ([]models.Rental, error) {
	return m.findAllResult, m.findAllErr
}

func (m *mockRentalRepo) Update(_ context.Context, _ *models.Rental) error {
	return m.updateErr
}

// --- CreateRental Tests ---

func TestCreateRental_Success(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{
			activeByCycleErr: gorm.ErrRecordNotFound,
			activeByUserErr:  gorm.ErrRecordNotFound,
		},
		&mockBikeRepo{existsResult: true},
	)

	rental, err := svc.CreateRental(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rental.Status != models.StatusActive {
		t.Errorf("expected status active, got %s", rental.Status)
	}
	if rental.StartTime.IsZero() {
		t.Error("expected start_time to be set")
	}
}

func TestCreateRental_BikeNotFound(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{},
		&mockBikeRepo{existsResult: false},
	)

	_, err := svc.CreateRental(context.Background(), uuid.New(), uuid.New())
	if err != ErrBikeNotFound {
		t.Errorf("expected ErrBikeNotFound, got %v", err)
	}
}

func TestCreateRental_BikeAlreadyRented(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{
			activeByCycleResult: &models.Rental{},
			activeByCycleErr:    nil,
		},
		&mockBikeRepo{existsResult: true},
	)

	_, err := svc.CreateRental(context.Background(), uuid.New(), uuid.New())
	if err != ErrBikeAlreadyRented {
		t.Errorf("expected ErrBikeAlreadyRented, got %v", err)
	}
}

func TestCreateRental_UserHasActiveRental(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{
			activeByCycleErr:   gorm.ErrRecordNotFound,
			activeByUserResult: &models.Rental{},
			activeByUserErr:    nil,
		},
		&mockBikeRepo{existsResult: true},
	)

	_, err := svc.CreateRental(context.Background(), uuid.New(), uuid.New())
	if err != ErrUserHasActive {
		t.Errorf("expected ErrUserHasActive, got %v", err)
	}
}

// --- FinalizeRental Tests ---

func TestFinalizeRental_Success(t *testing.T) {
	userID := uuid.New()
	rentalID := uuid.New()
	bikeID := uuid.New()

	rental := &models.Rental{
		ID:        rentalID,
		UserID:    userID,
		BicycleID: bikeID,
		Status:    models.StatusActive,
		StartTime: time.Now().UTC().Add(-1 * time.Hour),
	}

	svc := NewRentalService(
		&mockRentalRepo{findByIDResult: rental},
		&mockBikeRepo{},
	)

	result, err := svc.FinalizeRental(context.Background(), rentalID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != models.StatusFinalized {
		t.Errorf("expected status finalized, got %s", result.Status)
	}
	if result.EndTime == nil {
		t.Error("expected end_time to be set")
	}
	if result.DurationSeconds == nil || *result.DurationSeconds < 3500 {
		t.Error("expected duration_seconds to be approximately 3600")
	}
}

func TestFinalizeRental_NotFound(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{findByIDErr: gorm.ErrRecordNotFound},
		&mockBikeRepo{},
	)

	_, err := svc.FinalizeRental(context.Background(), uuid.New(), uuid.New())
	if err != ErrRentalNotFound {
		t.Errorf("expected ErrRentalNotFound, got %v", err)
	}
}

func TestFinalizeRental_NotOwner(t *testing.T) {
	rental := &models.Rental{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Status: models.StatusActive,
	}

	svc := NewRentalService(
		&mockRentalRepo{findByIDResult: rental},
		&mockBikeRepo{},
	)

	_, err := svc.FinalizeRental(context.Background(), rental.ID, uuid.New())
	if err != ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestFinalizeRental_AlreadyFinalized(t *testing.T) {
	userID := uuid.New()
	rental := &models.Rental{
		ID:     uuid.New(),
		UserID: userID,
		Status: models.StatusFinalized,
	}

	svc := NewRentalService(
		&mockRentalRepo{findByIDResult: rental},
		&mockBikeRepo{},
	)

	_, err := svc.FinalizeRental(context.Background(), rental.ID, userID)
	if err != ErrNotActive {
		t.Errorf("expected ErrNotActive, got %v", err)
	}
}

// --- GetActiveRental Tests ---

func TestGetActiveRental_Found(t *testing.T) {
	userID := uuid.New()
	rental := &models.Rental{
		ID:     uuid.New(),
		UserID: userID,
		Status: models.StatusActive,
	}

	svc := NewRentalService(
		&mockRentalRepo{activeByUserResult: rental},
		&mockBikeRepo{},
	)

	result, err := svc.GetActiveRental(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != rental.ID {
		t.Errorf("expected rental ID %s, got %s", rental.ID, result.ID)
	}
}

func TestGetActiveRental_NotFound(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{activeByUserErr: gorm.ErrRecordNotFound},
		&mockBikeRepo{},
	)

	_, err := svc.GetActiveRental(context.Background(), uuid.New())
	if err != ErrRentalNotFound {
		t.Errorf("expected ErrRentalNotFound, got %v", err)
	}
}

// --- ListAllRentals Tests ---

func TestListAllRentals_Success(t *testing.T) {
	rentals := []models.Rental{
		{ID: uuid.New(), Status: models.StatusActive},
		{ID: uuid.New(), Status: models.StatusFinalized},
	}

	svc := NewRentalService(
		&mockRentalRepo{findAllResult: rentals},
		&mockBikeRepo{},
	)

	result, err := svc.ListAllRentals(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 rentals, got %d", len(result))
	}
}

func TestListAllRentals_Empty(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{findAllResult: []models.Rental{}},
		&mockBikeRepo{},
	)

	result, err := svc.ListAllRentals(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 rentals, got %d", len(result))
	}
}

// --- CancelRental Tests ---

func TestCancelRental_Success(t *testing.T) {
	rental := &models.Rental{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		BicycleID: uuid.New(),
		Status:    models.StatusActive,
		StartTime: time.Now().UTC().Add(-1 * time.Hour),
	}

	svc := NewRentalService(
		&mockRentalRepo{findByIDResult: rental},
		&mockBikeRepo{},
	)

	result, err := svc.CancelRental(context.Background(), rental.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != models.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", result.Status)
	}
	if result.EndTime == nil {
		t.Error("expected end_time to be set")
	}
}

func TestCancelRental_NotFound(t *testing.T) {
	svc := NewRentalService(
		&mockRentalRepo{findByIDErr: gorm.ErrRecordNotFound},
		&mockBikeRepo{},
	)

	_, err := svc.CancelRental(context.Background(), uuid.New())
	if err != ErrRentalNotFound {
		t.Errorf("expected ErrRentalNotFound, got %v", err)
	}
}

func TestCancelRental_AlreadyFinalized(t *testing.T) {
	rental := &models.Rental{
		ID:     uuid.New(),
		Status: models.StatusFinalized,
	}

	svc := NewRentalService(
		&mockRentalRepo{findByIDResult: rental},
		&mockBikeRepo{},
	)

	_, err := svc.CancelRental(context.Background(), rental.ID)
	if err != ErrNotActive {
		t.Errorf("expected ErrNotActive, got %v", err)
	}
}
