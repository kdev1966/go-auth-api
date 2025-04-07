// controllers/user.go

package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetAllUsers récupère tous les utilisateurs (pour un usage admin, par exemple).
func GetAllUsers(c *gin.Context) {
	// Récupérer le rôle depuis le contexte
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès refusé : réservé aux administrateurs"})
		return
	}

	var users []models.User

	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des utilisateurs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserByID récupère un utilisateur selon l'ID passé en paramètre.
func GetUserByID(c *gin.Context) {
	// Extraction de l'ID depuis les paramètres
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Récupération des infos depuis le token (middleware)
	tokenUsername, _ := c.Get("username")
	tokenRole, _ := c.Get("role")

	// Vérification d'accès : admin ou propriétaire du compte
	if tokenRole != "admin" && tokenUsername != user.Username {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès refusé"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}


// UpdateUser met à jour les informations d'un utilisateur.
// Seul l'utilisateur lui-même peut mettre à jour son profil.
func UpdateUser(c *gin.Context) {
	// Paramètre ID
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// Récupération des infos du token
	tokenUserIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	tokenRole, _ := c.Get("role")

	tokenUserID, ok := tokenUserIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur d'identification de l'utilisateur"})
		return
	}

	// Autorisation : admin ou le bon utilisateur
	if tokenRole != "admin" && tokenUserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès refusé"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Champs à mettre à jour
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de hachage du mot de passe"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur mis à jour avec succès", "user": user})
}


// DeleteUser supprime un utilisateur.
// Seul l'utilisateur lui-même ou un admin peut réaliser cette opération.
func DeleteUser(c *gin.Context) {
	// Paramètre ID
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// Infos du token
	tokenUserIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	tokenRole, _ := c.Get("role")

	tokenUserID, ok := tokenUserIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur d'identification de l'utilisateur"})
		return
	}

	// Autorisation
	if tokenRole != "admin" && tokenUserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès refusé"})
		return
	}

	// Suppression
	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur supprimé avec succès"})
}

// GetMe retourne les infos de l'utilisateur connecté se basant sur l'ID du token
func GetMe(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID utilisateur non trouvé dans le contexte"})
		return
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Type d'ID utilisateur invalide"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Masquer le mot de passe avant de renvoyer l'utilisateur
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

