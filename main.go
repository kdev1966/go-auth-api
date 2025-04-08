// main.go

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"     // nécessaire pour les fichiers générés par swag
	ginSwagger "github.com/swaggo/gin-swagger" // nécessaire pour les fichiers générés par swag

	"github.com/kdev1966/go-auth-api/config"
	_ "github.com/kdev1966/go-auth-api/docs" // nécessaire pour les fichiers générés par swag
	"github.com/kdev1966/go-auth-api/models"
	"github.com/kdev1966/go-auth-api/routes"
)

// @title           Go Auth API
// @version         1.0
// @description     API d'authentification en Go avec JWT, upload, logs, etc.
// @termsOfService  http://example.com/terms/

// @contact.name   Ton Nom
// @contact.url    http://example.com
// @contact.email  ton.email@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:4000
// @BasePath  /api

func main() {
	// Charger les variables d'environnement depuis le fichier .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erreur lors du chargement du fichier .env", err)
	}

	// Release mode pour Gin
	gin.SetMode(gin.ReleaseMode)

	// Connexion à la base de données
	config.ConnectDatabase()

	// Migration automatique du modèle
	if err := config.DB.AutoMigrate(&models.User{}, &models.ActivityLog{}); err != nil {
		log.Fatal("Erreur lors de la migration de la base de données:", err)
	}
	log.Println("Migration réussie pour le modèle User et ActivityLog.")

	// Configuration des routes via le package routes
	router := routes.SetupRoutes()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Servir le dossier des uploads
	router.Static("/uploads", "./uploads") // Dossier pour les fichiers uploadés

	// Récupérer le port à partir des variables d'environnement, avec un fallback à 4000
	port := os.Getenv("PORT")
	// Si le port n'est pas défini, utiliser une valeur par défaut
	if port == "" {
		port = "4000" // Valeur par défaut
	}

	// Démarrage du serveur
	log.Println("Démarrage du serveur sur le port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
