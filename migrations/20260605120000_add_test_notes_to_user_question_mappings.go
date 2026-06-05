package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.ExecContext(
				ctx,
				`ALTER TABLE "user_question_mappings" ADD COLUMN IF NOT EXISTS "test_notes" jsonb NOT NULL DEFAULT '[]'::jsonb`,
			)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.ExecContext(
				ctx,
				`ALTER TABLE "user_question_mappings" DROP COLUMN IF EXISTS "test_notes"`,
			)
			return err
		},
	)
}

