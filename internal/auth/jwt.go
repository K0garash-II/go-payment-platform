package auth

import (
	"errors"
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
		UserID:    userID,
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		Issuer:    "go-payment-platform",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString(secret)

	return tokenString, err
}

func ValidateAccessToken(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer("go-payment-platform"),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return claims, nil
}
