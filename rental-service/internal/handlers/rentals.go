package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/models"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/schemas"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/services"
)

type RentalsHandler struct {
	service *services.RentalsService
}

func NewRentalsHandler(service *services.RentalsService) *RentalsHandler {
	return &RentalsHandler{service: service}
}

func (h *RentalsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.CreateRental)
	rg.PUT("/:id/finalize", h.FinalizeRental)
	rg.GET("/active", h.GetActiveRental)
}

func (h *RentalsHandler) CreateRental(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid user ID"})
		return
	}

	var req schemas.CreateRentalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "bicycle_id is required and must be a valid UUID"})
		return
	}

	rental, err := h.service.CreateRental(userID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toRentalResponse(rental))
}

func (h *RentalsHandler) FinalizeRental(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid user ID"})
		return
	}

	rentalIDStr := c.Param("id")
	rentalID, err := uuid.Parse(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid rental ID"})
		return
	}

	rental, err := h.service.FinalizeRental(userID, rentalID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRentalResponse(rental))
}

func (h *RentalsHandler) GetActiveRental(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid user ID"})
		return
	}

	resp, err := h.service.GetActiveRental(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func handleServiceError(c *gin.Context, err error) {
	switch err.(type) {
	case *services.ConflictError:
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
	case *services.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
	case *services.ForbiddenError:
		c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}

func toRentalResponse(r *models.Rental) schemas.RentalResponse {
	return schemas.RentalResponse{
		ID:        r.ID,
		UserID:    r.UserID,
		BicycleID: r.BicycleID,
		Status:    r.Status,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
		Duration:  r.Duration,
	}
}
