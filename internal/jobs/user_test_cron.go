package jobs

import (
	"context"
	"log/slog"
	"time"

	usersmodel "ai_testing/internal/modules/users/model"

	"github.com/uptrace/bun"
)

func StartUserTestMappingCron(ctx context.Context, logger *slog.Logger, db *bun.DB) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var rows []usersmodel.UserTestMapping
				err := db.NewSelect().
					Model(&rows).
					Where("status IN (?)", bun.In([]string{"initialized", "submitted"})).
					Order("created_at DESC").
					Limit(200).
					Scan(ctx)
				if err != nil {
					logger.Error("cron_user_test_mapping_error", "error", err.Error())
					continue
				}
				logger.Info("cron_user_test_mapping_scan", "count", len(rows))
			}
		}
	}()
}

