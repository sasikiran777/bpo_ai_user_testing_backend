package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"ai_testing/internal/config"
	"ai_testing/internal/http/middleware"
	authbootstrap "ai_testing/internal/modules/auth/bootstrap"
	authmw "ai_testing/internal/modules/auth/middleware"
	testsbootstrap "ai_testing/internal/modules/tests/bootstrap"
)

func NewRouter(logger *slog.Logger, db *bun.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS([]string{"http://localhost:3500", "https://assess.bposolutionsgroup.com"}))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	api := r.Group("/api/v1")
	publicRouter := api.Group("")
	protectedRouter := api.Group("")
	protectedRouter.Use(authmw.JWTAuth(db))

	RegisterPubPublicRoutes(publicRouter, db, cfg)
	RegisterProtectedRoutes(protectedRouter, db, cfg)

	return r
}

func RegisterPubPublicRoutes(r *gin.RouterGroup, db *bun.DB, cfg *config.Config) {
	authbootstrap.Register(r.Group("/auth"), db)
}

func RegisterProtectedRoutes(r *gin.RouterGroup, db *bun.DB, cfg *config.Config) {
	testsbootstrap.Register(r.Group("/tests"), db, cfg)
}
