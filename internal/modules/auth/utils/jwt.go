package utils

import (
	"ai_testing/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func jwtSecret() []byte {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is not configured")
	}
	return []byte(cfg.JWTSecret)
}

func GenerateToken(
	email string,
) (string, error) {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is not configured")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": email,
			"exp":   time.Now().Add(time.Duration(cfg.JWTTTLMinutes) * time.Minute).Unix(),
		},
	)

	return token.SignedString([]byte(cfg.JWTSecret))
}
