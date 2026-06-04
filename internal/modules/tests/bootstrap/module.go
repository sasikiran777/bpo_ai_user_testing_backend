package bootstrap

import (
	"ai_testing/internal/modules/tests/handler"
	"ai_testing/internal/modules/tests/repository"
	"ai_testing/internal/modules/tests/routes"
	"ai_testing/internal/modules/tests/service"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func Register(
	router *gin.RouterGroup,
	db *bun.DB,
) {
	repo := repository.New(db)
	svc := service.New(repo)
	hdl := handler.New(svc)
	routes.RegisterTestRoutes(router, hdl)
}
