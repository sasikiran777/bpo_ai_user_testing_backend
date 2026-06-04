package middleware

import (
	"net/http"
	"strings"

	"ai_testing/internal/config"
	userRepository "ai_testing/internal/modules/users/repository"
	"ai_testing/internal/shared/helpers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
)

func JWTAuth(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(rawToken, &claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			cfg := config.Load()

			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || token == nil || !token.Valid {
			helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		emailRaw, ok := claims["email"]
		if !ok {
			helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		email, ok := emailRaw.(string)
		if !ok || strings.TrimSpace(email) == "" {
			helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		userRepo := userRepository.New(db)
		user, err := userRepo.GetUserByEmail(c.Request.Context(), email)
		if err != nil {
			helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("email", email)
		c.Next()
	}
}

// bearerToken extracts the JWT from Authorization (Bearer <token>) without altering token casing.
func bearerToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", false
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
		authHeader = strings.TrimSpace(authHeader[len(bearerPrefix):])
	}

	if authHeader == "" {
		return "", false
	}
	return authHeader, true
}
