// controller/logs.go

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
)

func GetActivityLogs(c *gin.Context) {
	var logs []models.ActivityLog
	if err := config.DB.Order("created_at desc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer les logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}
