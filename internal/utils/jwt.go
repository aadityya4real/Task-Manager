package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("mysecretkey")

func GenerateToken(userID int, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),

		// small improvements (still simple)
		"iat": time.Now().Unix(),
	})

	return token.SignedString(secretKey)
}
