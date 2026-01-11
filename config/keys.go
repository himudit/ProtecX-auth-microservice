package config

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var PrivateKey *rsa.PrivateKey
var PublicKey *rsa.PublicKey

// func LoadRSAKeys() {
// 	// Read private key
// 	privateBytes, err := os.ReadFile("internal/keys/private.pem")
// 	if err != nil {
// 		log.Fatal("❌ Cannot read private key:", err)
// 	}

// 	// Read public key
// 	publicBytes, err := os.ReadFile("internal/keys/public.pem")
// 	if err != nil {
// 		log.Fatal("❌ Cannot read public key:", err)
// 	}

// 	// Parse private key → assign to package-level variable
// 	PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
// 	if err != nil {
// 		log.Fatal("❌ Invalid private key:", err)
// 	}

// 	// Parse public key → assign to package-level variable
// 	PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
// 	if err != nil {
// 		log.Fatal("❌ Invalid public key:", err)
// 	}

// 	log.Println("🔐 RSA keys loaded successfully")
// }

func LoadRSAKeys() {
	if os.Getenv("ENV") == "development" {
		// load from files
		privateBytes, _ := os.ReadFile("internal/keys/private.pem")
		publicBytes, _ := os.ReadFile("internal/keys/public.pem")

		PrivateKey, _ = jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
		PublicKey, _ = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
		return
	}

	// production → env
	privateKey := os.Getenv("JWT_PRIVATE_KEY")
	publicKey := os.Getenv("JWT_PUBLIC_KEY")

	PrivateKey, _ = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	PublicKey, _ = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
}
