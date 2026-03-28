package router

import (
	"github.com/bicycle-network/rental-service/internal/handler"
	"github.com/bicycle-network/rental-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(
	rentalHandler *handler.RentalHandler,
	healthHandler *handler.HealthHandler,
	jwtSecret string,
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(jwtSecret))
	{
		auth.POST("/rentals", rentalHandler.CreateRental)
		auth.PATCH("/rentals/:id/finalize", rentalHandler.FinalizeRental)
		auth.GET("/rentals/active", rentalHandler.GetActiveRental)
	}

	return r
}
