// controllers/user.go

package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
	"github.com/kdev1966/go-auth-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetAllUsers récupère tous les utilisateurs (pour un usage admin, par exemple).
func GetAllUsers(c *gin.Context) {
	// Vérifie si l'utilisateur est admin
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit"})
		return
	}

	// Récupère les paramètres de pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var users []models.User
	query := config.DB.Model(&models.User{})

	// Recherche (username ou email)
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	// Pagination
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des utilisateurs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       users,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": int((total + int64(limit) - 1) / int64(limit)),
	})
}

// GetUserByID godoc
// @Summary      Obtenir un utilisateur par son ID
// @Description  Accessible par un admin ou l'utilisateur lui-même
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      string  true  "ID de l'utilisateur"
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
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
	// Log de l'activité
	utils.LogActivity(uint(userID), "delete_account", "Suppression du compte par l'utilisateur ou un admin")

}

// HardDeleteUser supprime définitivement un utilisateur de la base de données.
// Seul un admin peut réaliser cette opération.
func HardDeleteUser(c *gin.Context) {
	// Récupération de l'ID depuis les paramètres
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// Récupération du rôle depuis le contexte
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès refusé : réservé aux administrateurs"})
		return
	}

	// Suppression définitive de l'utilisateur
	if err := config.DB.Unscoped().Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression définitive de l'utilisateur"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur supprimé définitivement avec succès"})
}

// RestoreUser restaure un utilisateur supprimé (soft delete).
// Seul un admin peut réaliser cette opération.
func RestoreUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// Vérifie que l'utilisateur est admin
	tokenRole, _ := c.Get("role")
	if tokenRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seul un administrateur peut restaurer un utilisateur"})
		return
	}

	// Mise à jour : suppression du deleted_at
	if err := config.DB.Unscoped().
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("deleted_at", nil).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Utilisateur restauré avec succès"})
}

// GetMe godoc
// @Summary      Retourne le profil utilisateur
// @Description  Donne les infos de l'utilisateur connecté
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Router       /me [get]
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

// UploadAvatar godoc
// @Summary      Upload de l'avatar de l'utilisateur
// @Description  Permet à un utilisateur connecté d’envoyer une image de profil
// @Tags         users
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Param        avatar  formData  file  true  "Fichier image à uploader"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /users/avatar [post]
func UploadAvatar(c *gin.Context) {
	// Récupère l'ID utilisateur depuis le token
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	userID := userIDRaw.(uint)

	// Charger l'utilisateur pour accéder à l'avatar actuel
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Utilisateur introuvable"})
		return
	}

	// Récupère le fichier envoyé dans le champ "avatar"
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier manquant"})
		return
	}

	// Si un avatar existait déjà, le supprimer du serveur
	if user.Avatar != "" {
		// Supposons que user.Avatar contienne un chemin relatif comme "/uploads/avatars/xxx.jpg"
		oldAvatarPath := user.Avatar
		// Supprimer le "/" initial si présent
		if len(oldAvatarPath) > 0 && oldAvatarPath[0] == '/' {
			oldAvatarPath = oldAvatarPath[1:]
		}
		// Tenter de supprimer le fichier (on peut logger l'erreur sans bloquer le processus)
		if err := os.Remove(oldAvatarPath); err != nil {
			// Log éventuel de l'erreur, mais ne pas interrompre l'upload
			// fmt.Printf("Erreur lors de la suppression de l'ancien avatar: %v\n", err)
		}
	}

	// Générer un nom de fichier unique pour éviter les collisions
	filename := fmt.Sprintf("user_%d_%s", userID, file.Filename)
	newFilePath := fmt.Sprintf("uploads/avatars/%s", filename)

	// Sauvegarde du nouveau fichier sur le disque
	if err := c.SaveUploadedFile(file, newFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de sauvegarde de l'avatar"})
		return
	}

	// Mettre à jour le champ Avatar dans la base de données
	avatarURL := "/" + newFilePath // par exemple, pour servir via router.Static("/uploads", "./uploads")
	if err := config.DB.Model(&user).Update("avatar", avatarURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de l'utilisateur"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Avatar mis à jour avec succès",
		"avatarUrl": avatarURL,
	})
	// Log de l'activité
	utils.LogActivity(userID, "update_avatar", "Mise à jour de l'avatar par l'utilisateur")
}
