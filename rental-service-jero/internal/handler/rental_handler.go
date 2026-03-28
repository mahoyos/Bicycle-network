package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bicycle-network/rental-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createRentalRequest struct {
	BicycleID string `json:"bicycle_id" binding:"required"`
}

type RentalHandler struct {
	svc service.RentalService
}

func NewRentalHandler(svc service.RentalService) *RentalHandler {
	return &RentalHandler{svc: svc}
}

func (h *RentalHandler) CreateRental(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createRentalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bicycle_id is required"})
		return
	}

	bicycleID, err := uuid.Parse(req.BicycleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bicycle_id format"})
		return
	}

	rental, err := h.svc.CreateRental(c.Request.Context(), userID, bicycleID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBikeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrBikeAlreadyRented):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrUserHasActive):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, rental)
}

func (h *RentalHandler) FinalizeRental(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	rentalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rental id format"})
		return
	}

	rental, err := h.svc.FinalizeRental(c.Request.Context(), rentalID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRentalNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotActive):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, rental)
}

func (h *RentalHandler) GetActiveRental(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	rental, err := h.svc.GetActiveRental(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrRentalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active rental found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, rental)
}

func (h *RentalHandler) ListAllRentals(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rentals, err := h.svc.ListAllRentals(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, rentals)
}

func (h *RentalHandler) CancelRental(c *gin.Context) {
	rentalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rental id format"})
		return
	}

	rental, err := h.svc.CancelRental(c.Request.Context(), rentalID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRentalNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotActive):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, rental)
}
