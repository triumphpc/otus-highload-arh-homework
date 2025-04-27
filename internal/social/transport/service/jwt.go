package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"otus-highload-arh-homework/internal/social/transport"
)

type JWTGenerator struct {
	secretKey []byte
	expiresIn time.Duration
}

func NewJWTGenerator(secret string, expiresIn time.Duration) *JWTGenerator {
	return &JWTGenerator{
		secretKey: []byte(secret),
		expiresIn: expiresIn,
	}
}

func (j *JWTGenerator) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(j.expiresIn).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTGenerator) ValidateToken(tokenStr string) (int, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, transport.ErrInvalidToken
	}

	userID, ok := claims["sub"].(float64) // JWT числа всегда float64
	if !ok {
		return 0, transport.ErrInvalidToken
	}

	return int(userID), nil
}
