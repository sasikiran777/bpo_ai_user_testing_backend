package bootstrap

import (
	"ai_testing/internal/config"
	"ai_testing/internal/modules/tests/handler"
	"ai_testing/internal/modules/tests/repository"
	"ai_testing/internal/modules/tests/routes"
	"ai_testing/internal/modules/tests/service"
	"ai_testing/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func Register(
	router *gin.RouterGroup,
	db *bun.DB,
	cfg *config.Config,
) {
	repo := repository.New(db)
	audioStore := storage.NewAudioStore(*cfg)
	svc := service.New(repo, audioStore)
	hdl := handler.New(svc, audioStore)
	routes.RegisterTestRoutes(router, hdl)
}
