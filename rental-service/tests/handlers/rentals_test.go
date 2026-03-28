package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/dependencies"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/handlers"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/models"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/schemas"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/services"
	"github.com/mahoyos/Bicycle-network/rental-service/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// Mock repo for handler tests
type mockRepo struct {
	rentals map[uuid.UUID]*models.Rental
}

func newMockRepo() *mockRepo {
	return &mockRepo{rentals: make(map[uuid.UUID]*models.Rental)}
}

func (m *mockRepo) Create(userID, bicycleID uuid.UUID) (*models.Rental, error) {
	r := &models.Rental{
		ID: uuid.New(), UserID: userID, BicycleID: bicycleID,
		Status: "active", StartTime: time.Now(),
	}
	m.rentals[r.ID] = r
	return r, nil
}

func (m *mockRepo) FindActiveByUserID(userID uuid.UUID) (*models.Rental, error) {
	for _, r := range m.rentals {
		if r.UserID == userID && r.Status == "active" {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) FindActiveByBicycleID(bicycleID uuid.UUID) (*models.Rental, error) {
	for _, r := range m.rentals {
		if r.BicycleID == bicycleID && r.Status == "active" {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) Finalize(rentalID uuid.UUID) (*models.Rental, error) {
	r, ok := m.rentals[rentalID]
	if !ok {
		return nil, nil
	}
	now := time.Now()
	dur := "0 hours 30 minutes 0 seconds"
	r.Status = "finalized"
	r.EndTime = &now
	r.Duration = &dur
	return r, nil
}

func (m *mockRepo) GetByID(rentalID uuid.UUID) (*models.Rental, error) {
	r, ok := m.rentals[rentalID]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (m *mockRepo) CreatePendingDelete(bicycleID uuid.UUID) error {
	return nil
}
func (m *mockRepo) FindPendingDeleteByBicycleID(bicycleID uuid.UUID) (*models.PendingDelete, error) {
	return nil, nil
}
func (m *mockRepo) MarkPendingDeleteProcessed(bicycleID uuid.UUID) error {
	return nil
}

type mockPub struct{}

func (m *mockPub) PublishBicycleReturned(bicycleID string) error { return nil }

func setupRouter(repo *mockRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := services.NewRentalsService(repo, &mockPub{})
	h := handlers.NewRentalsHandler(svc)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "rental-service"})
	})

	rentals := router.Group("/rentals")
	rentals.Use(dependencies.AuthMiddleware(helpers.PublicPEM, false))
	h.RegisterRoutes(rentals)

	return router
}

func TestHealthCheck(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "rental-service", body["service"])
}

func TestCreateRentalSuccess(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	userID := uuid.New().String()
	token := helpers.CreateToken(userID, "user", false)
	bicycleID := uuid.New()

	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "active", resp["status"])
}

func TestCreateRentalNoToken(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	bicycleID := uuid.New()
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestCreateRentalExpiredToken(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	token := helpers.CreateToken(uuid.New().String(), "user", true)
	bicycleID := uuid.New()

	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestCreateRentalInvalidToken(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	bicycleID := uuid.New()
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}

func TestCreateRentalMissingBicycleID(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	token := helpers.CreateToken(uuid.New().String(), "user", false)

	body := []byte(`{}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreateRentalUserHasActiveRental(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	userID := uuid.New().String()
	token := helpers.CreateToken(userID, "user", false)
	bicycleID := uuid.New()

	// Create first rental
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// Try to create second rental
	bicycleID2 := uuid.New()
	body2, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID2})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 409, w2.Code)
}

func TestCreateRentalBicycleAlreadyRented(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	bicycleID := uuid.New()

	// User 1 rents the bike
	user1 := uuid.New().String()
	token1 := helpers.CreateToken(user1, "user", false)
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// User 2 tries to rent the same bike
	user2 := uuid.New().String()
	token2 := helpers.CreateToken(user2, "user", false)
	body2, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body2))
	req2.Header.Set("Authorization", "Bearer "+token2)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 409, w2.Code)
}

func TestFinalizeRentalSuccess(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	userID := uuid.New().String()
	token := helpers.CreateToken(userID, "user", false)
	bicycleID := uuid.New()

	// Create rental first
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	rentalID := created["id"].(string)

	// Finalize
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/rentals/"+rentalID+"/finalize", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, "finalized", resp["status"])
	assert.NotNil(t, resp["end_time"])
	assert.NotNil(t, resp["duration"])
}

func TestFinalizeRentalNotFound(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	token := helpers.CreateToken(uuid.New().String(), "user", false)
	fakeID := uuid.New().String()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/rentals/"+fakeID+"/finalize", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestFinalizeRentalNotOwnedByUser(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	ownerID := uuid.New().String()
	ownerToken := helpers.CreateToken(ownerID, "user", false)
	bicycleID := uuid.New()

	// Owner creates rental
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	rentalID := created["id"].(string)

	// Other user tries to finalize
	otherToken := helpers.CreateToken(uuid.New().String(), "user", false)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/rentals/"+rentalID+"/finalize", nil)
	req2.Header.Set("Authorization", "Bearer "+otherToken)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 403, w2.Code)
}

func TestFinalizeRentalAlreadyFinalized(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	userID := uuid.New().String()
	token := helpers.CreateToken(userID, "user", false)
	bicycleID := uuid.New()

	// Create and finalize
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	rentalID := created["id"].(string)

	// First finalize
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/rentals/"+rentalID+"/finalize", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Second finalize attempt
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("PUT", "/rentals/"+rentalID+"/finalize", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w3, req3)

	assert.Equal(t, 409, w3.Code)
}

func TestFinalizeRentalNoToken(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/rentals/"+uuid.New().String()+"/finalize", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestGetActiveRentalSuccess(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	userID := uuid.New().String()
	token := helpers.CreateToken(userID, "user", false)
	bicycleID := uuid.New()

	// Create rental
	body, _ := json.Marshal(schemas.CreateRentalRequest{BicycleID: bicycleID})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// Get active
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/rentals/active", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, "active", resp["status"])
	assert.NotEmpty(t, resp["duration_so_far"])
}

func TestGetActiveRentalNoneFound(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	token := helpers.CreateToken(uuid.New().String(), "user", false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestGetActiveRentalNoToken(t *testing.T) {
	repo := newMockRepo()
	router := setupRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
