package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS middle_name text`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS country_code text NOT NULL DEFAULT ''`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS type_of_position_desired text NOT NULL DEFAULT ''`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS exp_in_months integer NOT NULL DEFAULT 0`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `UPDATE users SET phone = '' WHERE phone IS NULL`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN phone SET DEFAULT ''`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN phone SET NOT NULL`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
DO $$
BEGIN
	IF to_regclass('public.user_profiles') IS NOT NULL THEN
		UPDATE users u
		SET exp_in_months = COALESCE(up.total_exp_months, 0)
		FROM user_profiles up
		WHERE up.user_id = u.id;
	END IF;
END $$;
`); err != nil {
				return err
			}

			_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS user_profiles`)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS user_profiles (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	user_id uuid NOT NULL UNIQUE,
	total_exp_months integer NOT NULL DEFAULT 0,
	skills jsonb NOT NULL DEFAULT '[]'::jsonb,
	past_job_title text,
	company text,
	CONSTRAINT fk_user_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS middle_name`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS country_code`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS type_of_position_desired`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS exp_in_months`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN phone DROP NOT NULL`); err != nil {
				return err
			}
			_, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN phone DROP DEFAULT`)
			return err
		},
	)
}

