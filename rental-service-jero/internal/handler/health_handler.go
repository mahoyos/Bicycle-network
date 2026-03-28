package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db          *gorm.DB
	mqConnected func() bool
}

func NewHealthHandler(db *gorm.DB, mqConnected func() bool) *HealthHandler {
	return &HealthHandler{db: db, mqConnected: mqConnected}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "rental-service",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "database connection error",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "database not connected",
		})
		return
	}

	if !h.mqConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "rabbitmq not connected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
