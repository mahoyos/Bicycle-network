package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bicycle-network/rental-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func TestSetup_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rentalHandler := handler.NewRentalHandler(nil)
	healthHandler := handler.NewHealthHandler(nil, func() bool { return true })

	r := Setup(rentalHandler, healthHandler, "test-secret")

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSetup_RentalsRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rentalHandler := handler.NewRentalHandler(nil)
	healthHandler := handler.NewHealthHandler(nil, func() bool { return true })

	r := Setup(rentalHandler, healthHandler, "test-secret")

	req, _ := http.NewRequest("GET", "/rentals/active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 401 because no auth header
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSetup_PostRentalsRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rentalHandler := handler.NewRentalHandler(nil)
	healthHandler := handler.NewHealthHandler(nil, func() bool { return true })

	r := Setup(rentalHandler, healthHandler, "test-secret")

	req, _ := http.NewRequest("POST", "/rentals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
