// controllers/auth.go
package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
	"github.com/kdev1966/go-auth-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims étend les réclamations JWT en ajoutant l'ID de l'utilisateur.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// generateJWT génère un token JWT en incluant l'ID de l'utilisateur.
func GenerateToken(user models.User) (string, error) {
	mySigningKey := []byte(os.Getenv("JWT_SECRET"))

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Expiration 24h
			Issuer:    "go-auth-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Register crée un nouvel utilisateur.
func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"` // Optionnel, par défaut "user"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hachage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du hachage du mot de passe"})
		return
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Utilisateur créé avec succès"})
}

// Login authentifie l'utilisateur et génère un token JWT.
func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non trouvé"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Comparaison des mots de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe incorrect"})
		return
	}

	// Générer le token JWT
	accessToken, err := utils.GenerateToken(user.ID, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, 7*24*time.Hour)

	// Sauvegarder le refresh token côté DB
	config.DB.Model(&user).Update("refresh_token", refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Connexion réussie",
		"role":          user.Role,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	utils.LogActivity(user.ID, "login", "Utilisateur connecté avec succès")
}

// refresh 	token
func RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token manquant"})
		return
	}

	// Vérifie le token
	token, err := jwt.Parse(body.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token invalide"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Compare avec celui en base
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil || user.RefreshToken != body.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token non reconnu"})
		return
	}

	// Génère un nouveau access token
	newAccessToken, err := utils.GenerateToken(userID, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de générer un nouveau token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
