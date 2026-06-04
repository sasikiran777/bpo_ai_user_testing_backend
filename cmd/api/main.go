package main

import (
	"context"
	"errors"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai_testing/internal/config"
	database "ai_testing/internal/db"
	apihttp "ai_testing/internal/http"
	"ai_testing/internal/log"

	"github.com/uptrace/bun"
)

func main() {
	cfg := config.Load()
	logger := log.New(cfg)
	db := mustDB(&cfg)

	router := apihttp.NewRouter(logger, db, &cfg)

	srv := &nethttp.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server_listen", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			logger.Error("server_error", "error", err.Error())
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
	logger.Info("server_shutdown")
}

func mustDB(cfg *config.Config) *bun.DB {
	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	return db
}
