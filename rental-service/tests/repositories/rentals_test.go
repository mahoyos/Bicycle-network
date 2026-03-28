package repositories_test

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return db, mock, sqlDB
}

var rentalCols = []string{"id", "user_id", "bicycle_id", "status", "start_time", "end_time", "duration"}

func TestCreateInsertsRentalWithActiveStatus(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewRentalsRepository(db)
	userID := uuid.New()
	bicycleID := uuid.New()
	rentalID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "start_time"}).
		AddRow(rentalID, now)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rentals"`)).
		WillReturnRows(rows)
	mock.ExpectCommit()

	rental, err := repo.Create(userID, bicycleID)
	assert.NoError(t, err)
	assert.NotNil(t, rental)
	assert.Equal(t, "active", rental.Status)
	assert.Equal(t, userID, rental.UserID)
	assert.Equal(t, bicycleID, rental.BicycleID)
}

func TestFindActiveByUserIDReturnsActiveOnly(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewRentalsRepository(db)
	userID := uuid.New()
	rentalID := uuid.New()
	bicycleID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(rentalCols).
		AddRow(rentalID, userID, bicycleID, "active", now, nil, nil)

	// GORM First() adds ORDER BY id LIMIT 1, so we match with AnyArg for the LIMIT
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "rentals" WHERE user_id = $1 AND status = 'active'`)).
		WithArgs(userID, 1). // userID + LIMIT 1
		WillReturnRows(rows)

	rental, err := repo.FindActiveByUserID(userID)
	assert.NoError(t, err)
	assert.NotNil(t, rental)
	assert.Equal(t, "active", rental.Status)
}

func TestFindActiveByUserIDReturnsNilWhenNone(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewRentalsRepository(db)
	userID := uuid.New()

	rows := sqlmock.NewRows(rentalCols)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "rentals" WHERE user_id = $1 AND status = 'active'`)).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	rental, err := repo.FindActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Nil(t, rental)
}

func TestFindActiveByBicycleIDReturnsActiveRental(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewRentalsRepository(db)
	bicycleID := uuid.New()
	userID := uuid.New()
	rentalID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(rentalCols).
		AddRow(rentalID, userID, bicycleID, "active", now, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "rentals" WHERE bicycle_id = $1 AND status = 'active'`)).
		WithArgs(bicycleID, 1).
		WillReturnRows(rows)

	rental, err := repo.FindActiveByBicycleID(bicycleID)
	assert.NoError(t, err)
	assert.NotNil(t, rental)
	assert.Equal(t, bicycleID, rental.BicycleID)
}

func TestFindActiveByBicycleIDReturnsNilWhenFinalized(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewRentalsRepository(db)
	bicycleID := uuid.New()

	rows := sqlmock.NewRows(rentalCols)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "rentals" WHERE bicycle_id = $1 AND status = 'active'`)).
		WithArgs(bicycleID, 1).
		WillReturnRows(rows)

	rental, err := repo.FindActiveByBicycleID(bicycleID)
	assert.NoError(t, err)
	assert.Nil(t, rental)
}
