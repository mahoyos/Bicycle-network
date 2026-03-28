package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"


	"github.com/bicycle-network/rental-service/internal/models"
	"github.com/bicycle-network/rental-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mock RentalService ---

type mockRentalService struct {
	createResult   *models.Rental
	createErr      error
	finalizeResult *models.Rental
	finalizeErr    error
	activeResult   *models.Rental
	activeErr      error
}

func (m *mockRentalService) CreateRental(_ context.Context, _, _ uuid.UUID) (*models.Rental, error) {
	return m.createResult, m.createErr
}

func (m *mockRentalService) FinalizeRental(_ context.Context, _, _ uuid.UUID) (*models.Rental, error) {
	return m.finalizeResult, m.finalizeErr
}

func (m *mockRentalService) GetActiveRental(_ context.Context, _ uuid.UUID) (*models.Rental, error) {
	return m.activeResult, m.activeErr
}

func setupTestRouter(svc service.RentalService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewRentalHandler(svc)

	// Simulate auth middleware by setting user_id
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", uuid.MustParse("11111111-1111-1111-1111-111111111111"))
		c.Next()
	}

	r.POST("/rentals", authMiddleware, h.CreateRental)
	r.PATCH("/rentals/:id/finalize", authMiddleware, h.FinalizeRental)
	r.GET("/rentals/active", authMiddleware, h.GetActiveRental)

	return r
}

func TestCreateRental_Success(t *testing.T) {
	bikeID := uuid.New()
	rental := &models.Rental{
		ID:        uuid.New(),
		UserID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BicycleID: bikeID,
		Status:    models.StatusActive,
		StartTime: time.Now().UTC(),
	}

	router := setupTestRouter(&mockRentalService{createResult: rental})

	body, _ := json.Marshal(map[string]string{"bicycle_id": bikeID.String()})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateRental_MissingBicycleID(t *testing.T) {
	router := setupTestRouter(&mockRentalService{})

	body, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateRental_InvalidBicycleID(t *testing.T) {
	router := setupTestRouter(&mockRentalService{})

	body, _ := json.Marshal(map[string]string{"bicycle_id": "not-a-uuid"})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateRental_BikeNotFound(t *testing.T) {
	router := setupTestRouter(&mockRentalService{createErr: service.ErrBikeNotFound})

	body, _ := json.Marshal(map[string]string{"bicycle_id": uuid.New().String()})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCreateRental_Conflict(t *testing.T) {
	router := setupTestRouter(&mockRentalService{createErr: service.ErrBikeAlreadyRented})

	body, _ := json.Marshal(map[string]string{"bicycle_id": uuid.New().String()})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestFinalizeRental_Success(t *testing.T) {
	rentalID := uuid.New()
	now := time.Now().UTC()
	dur := 3600
	rental := &models.Rental{
		ID:              rentalID,
		UserID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BicycleID:       uuid.New(),
		Status:          models.StatusFinalized,
		StartTime:       now.Add(-1 * time.Hour),
		EndTime:         &now,
		DurationSeconds: &dur,
	}

	router := setupTestRouter(&mockRentalService{finalizeResult: rental})

	req, _ := http.NewRequest("PATCH", "/rentals/"+rentalID.String()+"/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFinalizeRental_InvalidID(t *testing.T) {
	router := setupTestRouter(&mockRentalService{})

	req, _ := http.NewRequest("PATCH", "/rentals/not-a-uuid/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFinalizeRental_NotFound(t *testing.T) {
	router := setupTestRouter(&mockRentalService{finalizeErr: service.ErrRentalNotFound})

	req, _ := http.NewRequest("PATCH", "/rentals/"+uuid.New().String()+"/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestFinalizeRental_Forbidden(t *testing.T) {
	router := setupTestRouter(&mockRentalService{finalizeErr: service.ErrNotOwner})

	req, _ := http.NewRequest("PATCH", "/rentals/"+uuid.New().String()+"/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGetActiveRental_Found(t *testing.T) {
	rental := &models.Rental{
		ID:        uuid.New(),
		UserID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BicycleID: uuid.New(),
		Status:    models.StatusActive,
		StartTime: time.Now().UTC(),
	}

	router := setupTestRouter(&mockRentalService{activeResult: rental})

	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetActiveRental_NotFound(t *testing.T) {
	router := setupTestRouter(&mockRentalService{activeErr: service.ErrRentalNotFound})

	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCreateRental_UserHasActiveConflict(t *testing.T) {
	router := setupTestRouter(&mockRentalService{createErr: service.ErrUserHasActive})

	body, _ := json.Marshal(map[string]string{"bicycle_id": uuid.New().String()})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestCreateRental_InternalError(t *testing.T) {
	router := setupTestRouter(&mockRentalService{createErr: context.DeadlineExceeded})

	body, _ := json.Marshal(map[string]string{"bicycle_id": uuid.New().String()})
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestFinalizeRental_NotActive(t *testing.T) {
	router := setupTestRouter(&mockRentalService{finalizeErr: service.ErrNotActive})

	req, _ := http.NewRequest("PATCH", "/rentals/"+uuid.New().String()+"/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestFinalizeRental_InternalError(t *testing.T) {
	router := setupTestRouter(&mockRentalService{finalizeErr: context.DeadlineExceeded})

	req, _ := http.NewRequest("PATCH", "/rentals/"+uuid.New().String()+"/finalize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetActiveRental_InternalError(t *testing.T) {
	router := setupTestRouter(&mockRentalService{activeErr: context.DeadlineExceeded})

	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
