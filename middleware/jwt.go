// middleware/jwt.go
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware vérifie la validité du token JWT et injecte les claims dans le contexte.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer le token depuis l'en-tête Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing"})
			c.Abort()
			return
		}

		// Vérifier que le token commence bien par "Bearer "
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token format is invalid"})
			c.Abort()
			return
		}

		mySigningKey := []byte(os.Getenv("JWT_SECRET"))
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Vérifier le type d'algorithme de signature
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return mySigningKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extraire les claims et les placer dans le contexte
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			// Extraction de l'ID utilisateur, en supposant qu'il est stocké comme un nombre
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", uint(userID))
			}
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}



