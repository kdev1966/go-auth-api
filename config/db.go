// config/db.go
// Package config provides database connection configuration for a Go application.
// It uses GORM with PostgreSQL as the database driver.

package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Exemple d'URL de connexion
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur lors de la connexion à la base de données:", err)
	}

		// Configuration du pool de connexions
		sqlDB, err := database.DB()
		if err != nil {
			log.Fatal("Erreur lors de l'obtention de la connexion SQL :", err)
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

	DB = database
}
