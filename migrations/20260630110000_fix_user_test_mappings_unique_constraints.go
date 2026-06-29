package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `ALTER TABLE "user_test_mappings" DROP CONSTRAINT IF EXISTS "user_test_mappings_user_id_key"`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE "user_test_mappings" DROP CONSTRAINT IF EXISTS "user_test_mappings_test_id_key"`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'user_test_mappings_user_id_test_id_key'
	) THEN
		ALTER TABLE "user_test_mappings"
			ADD CONSTRAINT "user_test_mappings_user_id_test_id_key" UNIQUE ("user_id", "test_id");
	END IF;
END $$;
`); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `ALTER TABLE "user_test_mappings" DROP CONSTRAINT IF EXISTS "user_test_mappings_user_id_test_id_key"`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'user_test_mappings_user_id_key'
	) THEN
		ALTER TABLE "user_test_mappings"
			ADD CONSTRAINT "user_test_mappings_user_id_key" UNIQUE ("user_id");
	END IF;
END $$;
`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'user_test_mappings_test_id_key'
	) THEN
		ALTER TABLE "user_test_mappings"
			ADD CONSTRAINT "user_test_mappings_test_id_key" UNIQUE ("test_id");
	END IF;
END $$;
`); err != nil {
				return err
			}
			return nil
		},
	)
}
