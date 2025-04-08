// utils/logger.go

package utils

import (
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
)

func LogActivity(userID uint, action, details string) {
	log := models.ActivityLog{
		UserID:  userID,
		Action:  action,
		Details: details,
	}
	config.DB.Create(&log)
}
