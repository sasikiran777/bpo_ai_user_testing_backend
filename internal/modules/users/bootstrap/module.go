package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"ai_testing/internal/modules/users/handler"
	"ai_testing/internal/modules/users/repository"
	"ai_testing/internal/modules/users/routes"
	"ai_testing/internal/modules/users/service"
)

func Register(api *gin.RouterGroup, db *bun.DB) {
	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	routes.Register(api, h)
}
