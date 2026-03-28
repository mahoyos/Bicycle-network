package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/models"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/schemas"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/services"
	"github.com/stretchr/testify/assert"
)

// Mock repository
type mockRepo struct {
	createFn              func(uuid.UUID, uuid.UUID) (*models.Rental, error)
	findActiveByUserIDFn  func(uuid.UUID) (*models.Rental, error)
	findActiveByBikeIDFn  func(uuid.UUID) (*models.Rental, error)
	finalizeFn            func(uuid.UUID) (*models.Rental, error)
	getByIDFn             func(uuid.UUID) (*models.Rental, error)
}

func (m *mockRepo) Create(userID, bicycleID uuid.UUID) (*models.Rental, error) {
	return m.createFn(userID, bicycleID)
}
func (m *mockRepo) FindActiveByUserID(userID uuid.UUID) (*models.Rental, error) {
	return m.findActiveByUserIDFn(userID)
}
func (m *mockRepo) FindActiveByBicycleID(bicycleID uuid.UUID) (*models.Rental, error) {
	return m.findActiveByBikeIDFn(bicycleID)
}
func (m *mockRepo) Finalize(rentalID uuid.UUID) (*models.Rental, error) {
	return m.finalizeFn(rentalID)
}
func (m *mockRepo) GetByID(rentalID uuid.UUID) (*models.Rental, error) {
	return m.getByIDFn(rentalID)
}

// Mock messaging
type mockPublisher struct {
	called   bool
	bikeID   string
	returnErr error
}

func (m *mockPublisher) PublishBicycleReturned(bicycleID string) error {
	m.called = true
	m.bikeID = bicycleID
	return m.returnErr
}

func TestCreateRentalChecksActiveRentalFirst(t *testing.T) {
	userID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{ID: uuid.New(), UserID: uid, Status: "active"}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.CreateRental(userID, schemas.CreateRentalRequest{BicycleID: bicycleID})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "User already has an active rental")
}

func TestCreateRentalRejectsBicycleAlreadyRented(t *testing.T) {
	userID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return nil, nil // no active rental for user
		},
		findActiveByBikeIDFn: func(bid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{ID: uuid.New(), BicycleID: bid, Status: "active"}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.CreateRental(userID, schemas.CreateRentalRequest{BicycleID: bicycleID})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bicycle is not available")
}

func TestCreateRentalSuccess(t *testing.T) {
	userID := uuid.New()
	bicycleID := uuid.New()
	rentalID := uuid.New()

	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return nil, nil
		},
		findActiveByBikeIDFn: func(bid uuid.UUID) (*models.Rental, error) {
			return nil, nil
		},
		createFn: func(uid, bid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: rentalID, UserID: uid, BicycleID: bid,
				Status: "active", StartTime: time.Now(),
			}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	rental, err := svc.CreateRental(userID, schemas.CreateRentalRequest{BicycleID: bicycleID})

	assert.NoError(t, err)
	assert.Equal(t, rentalID, rental.ID)
	assert.Equal(t, "active", rental.Status)
}

func TestCreateRentalAllowsBicycleAfterPreviousFinalized(t *testing.T) {
	userID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return nil, nil
		},
		findActiveByBikeIDFn: func(bid uuid.UUID) (*models.Rental, error) {
			return nil, nil // no active rental = bike is free
		},
		createFn: func(uid, bid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: uuid.New(), UserID: uid, BicycleID: bid,
				Status: "active", StartTime: time.Now(),
			}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	rental, err := svc.CreateRental(userID, schemas.CreateRentalRequest{BicycleID: bicycleID})

	assert.NoError(t, err)
	assert.NotNil(t, rental)
	assert.Equal(t, "active", rental.Status)
}

func TestPublishBicycleReturnedCalledAfterFinalize(t *testing.T) {
	userID := uuid.New()
	rentalID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		getByIDFn: func(rid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: rid, UserID: userID, BicycleID: bicycleID,
				Status: "active", StartTime: time.Now().Add(-1 * time.Hour),
			}, nil
		},
		finalizeFn: func(rid uuid.UUID) (*models.Rental, error) {
			now := time.Now()
			dur := "1 hours 0 minutes 0 seconds"
			return &models.Rental{
				ID: rid, UserID: userID, BicycleID: bicycleID,
				Status: "finalized", EndTime: &now, Duration: &dur,
			}, nil
		},
	}

	pub := &mockPublisher{}
	svc := services.NewRentalsService(repo, pub)
	_, err := svc.FinalizeRental(userID, rentalID)

	assert.NoError(t, err)
	assert.True(t, pub.called)
	assert.Equal(t, bicycleID.String(), pub.bikeID)
}

func TestPublishNotCalledWhenDBUpdateFails(t *testing.T) {
	userID := uuid.New()
	rentalID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		getByIDFn: func(rid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: rid, UserID: userID, BicycleID: bicycleID,
				Status: "active", StartTime: time.Now(),
			}, nil
		},
		finalizeFn: func(rid uuid.UUID) (*models.Rental, error) {
			return nil, errors.New("db error")
		},
	}

	pub := &mockPublisher{}
	svc := services.NewRentalsService(repo, pub)
	_, err := svc.FinalizeRental(userID, rentalID)

	assert.Error(t, err)
	assert.False(t, pub.called)
}

func TestFinalizeRentalNotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(rid uuid.UUID) (*models.Rental, error) {
			return nil, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.FinalizeRental(uuid.New(), uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rental not found")
}

func TestFinalizeRentalNotOwnedByUser(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	rentalID := uuid.New()

	repo := &mockRepo{
		getByIDFn: func(rid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: rid, UserID: ownerID, Status: "active",
				StartTime: time.Now(),
			}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.FinalizeRental(otherUserID, rentalID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rental does not belong to user")
}

func TestFinalizeRentalAlreadyFinalized(t *testing.T) {
	userID := uuid.New()
	rentalID := uuid.New()

	repo := &mockRepo{
		getByIDFn: func(rid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: rid, UserID: userID, Status: "finalized",
				StartTime: time.Now(),
			}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.FinalizeRental(userID, rentalID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rental is not active")
}

func TestGetActiveRentalSuccess(t *testing.T) {
	userID := uuid.New()
	bicycleID := uuid.New()

	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return &models.Rental{
				ID: uuid.New(), UserID: uid, BicycleID: bicycleID,
				Status: "active", StartTime: time.Now().Add(-30 * time.Minute),
			}, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	resp, err := svc.GetActiveRental(userID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "active", resp.Status)
	assert.NotEmpty(t, resp.DurationSoFar)
}

func TestGetActiveRentalNoneFound(t *testing.T) {
	repo := &mockRepo{
		findActiveByUserIDFn: func(uid uuid.UUID) (*models.Rental, error) {
			return nil, nil
		},
	}

	svc := services.NewRentalsService(repo, nil)
	_, err := svc.GetActiveRental(uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No active rental found")
}
