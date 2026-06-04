package migrations

import (
	"context"

	"github.com/uptrace/bun"

	usersmodel "ai_testing/internal/modules/users/model"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
				return err
			}

			if _, err := db.NewCreateTable().
				Model((*usersmodel.User)(nil)).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			_, err := db.NewCreateTable().
				Model((*usersmodel.UserProfile)(nil)).
				IfNotExists().
				ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`).
				Exec(ctx)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewDropTable().
				Model((*usersmodel.UserProfile)(nil)).
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			_, err := db.NewDropTable().
				Model((*usersmodel.User)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}
