// main.go

package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kdev1966/go-auth-api/config"
	"github.com/kdev1966/go-auth-api/models"
	"github.com/kdev1966/go-auth-api/routes"
)

func main() {
	// Charger les variables d'environnement depuis le fichier .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erreur lors du chargement du fichier .env", err)
	}

	// Connexion à la base de données
	config.ConnectDatabase()

	// Migration automatique du modèle
	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Erreur lors de la migration de la base de données:", err)
	}
	log.Println("Migration réussie pour le modèle User.")

	// Configuration des routes via le package routes
	router := routes.SetupRoutes()

	// Récupérer le port à partir des variables d'environnement, avec un fallback à 8080
	port := os.Getenv("PORT")
	// Si le port n'est pas défini, utiliser une valeur par défaut
	if port == "" {
		port = "4000" // Valeur par défaut
	}


	// Démarrage du serveur
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}





