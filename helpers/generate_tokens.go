package helpers

import (
	"chirpy/database"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GenerateRefreshToken(userId string) (database.RefreshToken, error) {
	// Refresh token will expire in 60 days
	expiredAt := time.Now().Add(24 * time.Hour * 60)

	ran := make([]byte, 32)
	_, err := rand.Read(ran)
	if err != nil {
		return database.RefreshToken{}, err
	}

	encodedStr := hex.EncodeToString(ran)

	refreshToken := database.RefreshToken{
		UserId:   userId,
		Token:    encodedStr,
		ExpireAt: expiredAt,
	}

	return refreshToken, nil
}

func GenerateJWTToken(userId string, secretKey string) (string, error) {
	mySigningKey := []byte(secretKey)

	// JWT will expire after 1 hour
	timeout := time.Now().UTC().Add(time.Hour * 1)

	claims := &jwt.RegisteredClaims{
		Issuer:   "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),

		Subject:   userId,
		ExpiresAt: jwt.NewNumericDate(timeout),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", fmt.Errorf("error while signing token: %s", err)
	}

	return ss, nil
}
