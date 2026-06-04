package migrations

import (
	"context"

	"github.com/uptrace/bun"

	testsmodel "ai_testing/internal/modules/tests/model"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
				return err
			}

			_, err := db.NewCreateTable().
				Model((*testsmodel.Test)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*testsmodel.Test)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}

