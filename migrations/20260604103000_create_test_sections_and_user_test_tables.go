package migrations

import (
	"context"

	"github.com/uptrace/bun"

	testsmodel "ai_testing/internal/modules/tests/model"
	usersmodel "ai_testing/internal/modules/users/model"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
				return err
			}

			if _, err := db.NewCreateTable().
				Model((*testsmodel.TestSectionMapping)(nil)).
				IfNotExists().
				ForeignKey(`("test_id") REFERENCES "tests" ("id") ON DELETE CASCADE`).
				Exec(ctx); err != nil {
				return err
			}

			if _, err := db.NewCreateTable().
				Model((*usersmodel.UserTestMapping)(nil)).
				IfNotExists().
				ForeignKey(`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`).
				ForeignKey(`("test_id") REFERENCES "tests" ("id") ON DELETE CASCADE`).
				Exec(ctx); err != nil {
				return err
			}

			if _, err := db.NewCreateTable().
				Model((*usersmodel.UserQuestionMapping)(nil)).
				IfNotExists().
				ForeignKey(`("user_test_mapping_id") REFERENCES "user_test_mappings" ("id") ON DELETE CASCADE`).
				ForeignKey(`("test_section_mapping_id") REFERENCES "test_section_mappings" ("id") ON DELETE CASCADE`).
				Exec(ctx); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, db *bun.DB) error {

			if _, err := db.NewDropTable().
				Model((*usersmodel.UserQuestionMapping)(nil)).
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			if _, err := db.NewDropTable().
				Model((*usersmodel.UserTestMapping)(nil)).
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			_, err := db.NewDropTable().
				Model((*testsmodel.TestSectionMapping)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}
