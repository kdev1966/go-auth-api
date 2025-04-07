// models/user.go

package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role	 string `gorm:"default:'user';not null"` // 'user' or 'admin'
}
