package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"ai_testing/internal/modules/auth/handler"
	"ai_testing/internal/modules/auth/routes"
	"ai_testing/internal/modules/auth/service"
	"ai_testing/internal/modules/users/repository"
)

func Register(
	router *gin.RouterGroup,
	db *bun.DB,
) {
	repo := repository.New(db)
	authSvc := service.New(repo)
	authHdl := handler.New(authSvc)
	routes.RegisterAuthRoutes(router, authHdl)
}
