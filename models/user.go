// models/user.go

package models

import (
	"time"
)

//	type User struct {
//		gorm.Model
//		Username     string `gorm:"unique;not null"`
//		Email        string `gorm:"unique;not null"`
//		Password     string `gorm:"not null"`
//		Role         string `gorm:"default:'user';not null"` // 'user' or 'admin'
//		Avatar       string `gorm:"type:text" json:"avatar"`
//		RefreshToken string `gorm:"type:text" json:"-"`
//	}
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Utilisation de *time.Time ici

	Username     string `gorm:"unique;not null" json:"username"`
	Email        string `gorm:"unique;not null" json:"email"`
	Password     string `gorm:"not null" json:"-"` // masqué dans les réponses JSON
	Role         string `gorm:"default:'user';not null" json:"role"`
	Avatar       string `gorm:"type:text" json:"avatar"`
	RefreshToken string `gorm:"type:text" json:"-"`
}
