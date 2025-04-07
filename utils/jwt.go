// utils/jwt.go

package utils

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Structure du payload du JWT
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// Fonction pour générer un JWT
func GenerateJWT(username, role string) (string, error) {
	// Définir la clé secrète pour signer le token
	var mySigningKey = []byte(os.Getenv("JWT_SECRET"))

	// Créer le payload du token
	claims := Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Le token expire après 24 heures
			Issuer:    "go-auth-api",
		},
	}

	// Créer un nouveau token à signer
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer le token avec la clé secrète
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
