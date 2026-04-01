package dependencies

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware(secretKey string, disabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if disabled {
			// Set a default user for development
			c.Set("user_id", "00000000-0000-0000-0000-000000000000")
			c.Set("user_role", "user")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid authorization header format"})
			return
		}

		tokenStr := parts[1]

		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Token has expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
			return
		}

		c.Set("user_id", claims.Sub)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}
