// middleware/is_admin.go
// Compare this snippet from routes/routes.go:

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
			return
		}
		c.Next()
	}
}
