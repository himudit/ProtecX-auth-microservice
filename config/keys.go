package config

import (
	"crypto/rsa"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var PrivateKey *rsa.PrivateKey
var PublicKey *rsa.PublicKey

func LoadRSAKeys() {
	if os.Getenv("ENV") == "development" {
		// load from files
		privateBytes, _ := os.ReadFile("internal/keys/private.pem")
		publicBytes, _ := os.ReadFile("internal/keys/public.pem")

		PrivateKey, _ = jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
		PublicKey, _ = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
		log.Println("🔐 RSA keys loaded successfully")
		return
	}

	// production → env
	privateKey := os.Getenv("JWT_PRIVATE_KEY")
	publicKey := os.Getenv("JWT_PUBLIC_KEY")

	PrivateKey, _ = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	PublicKey, _ = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	log.Println("🔐 RSA keys loaded successfully")
}
