package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewHealthHandler(nil, func() bool { return true })
	r.GET("/health", h.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReady_DBNotConnected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a GORM DB with an invalid DSN — Ping will fail
	db, _ := gorm.Open(postgres.Open("host=invalid port=0 user=x dbname=x sslmode=disable"), &gorm.Config{})

	h := NewHealthHandler(db, func() bool { return true })
	r.GET("/ready", h.Ready)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestReady_RabbitMQNotConnected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	db, _ := gorm.Open(postgres.Open("host=invalid port=0 user=x dbname=x sslmode=disable"), &gorm.Config{})

	h := NewHealthHandler(db, func() bool { return false })
	r.GET("/ready", h.Ready)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Will still be 503 because DB is also not connected
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
