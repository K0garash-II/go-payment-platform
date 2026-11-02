package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID int `json:"user_id"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID int, secret []byte) (string, error) {
	now := time.Now()
	
	claim := Claims{
		UserID: userID,
		Subject: strconv.Itoa(userID),
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		Issuer: "go-payment-platform",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString(secret)

	return tokenString, err
}