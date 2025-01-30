package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your_secret_key")

func GenerateJWT(userID, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtKey)
}

// GenerateVerificationToken generates a random verification token

func GenerateVerificationToken() (string, error) {

	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {

		return "", err

	}

	return hex.EncodeToString(bytes), nil

}
