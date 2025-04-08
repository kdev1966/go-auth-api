// models/activity_log.go

package models

import (
	"time"
)

type ActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Action    string    `json:"action"`  // ex: "login", "delete_account"
	Details   string    `json:"details"` // optionnel : "a supprimé son compte", "changé son mot de passe", etc.
	CreatedAt time.Time `json:"created_at"`
}
