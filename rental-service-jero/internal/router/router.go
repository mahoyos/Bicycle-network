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

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(jwtSecret))
	{
		auth.POST("/rentals", rentalHandler.CreateRental)
		auth.PATCH("/rentals/:id/finalize", rentalHandler.FinalizeRental)
		auth.GET("/rentals/active", rentalHandler.GetActiveRental)

		admin := auth.Group("/")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/rentals", rentalHandler.ListAllRentals)
			admin.PATCH("/rentals/:id/cancel", rentalHandler.CancelRental)
		}
	}

	return r
}
