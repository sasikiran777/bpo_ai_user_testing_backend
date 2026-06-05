package migrations

import (
	"context"

	testsmodel "ai_testing/internal/modules/tests/model"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewCreateTable().
				Model((*testsmodel.SpeakingTopic)(nil)).
				IfNotExists().
				ForeignKey(`("test_section_mapping_id") REFERENCES "test_section_mappings" ("id") ON DELETE CASCADE`).
				Exec(ctx); err != nil {
				return err
			}

			_, err := db.NewCreateTable().
				Model((*testsmodel.ReadingComprehension)(nil)).
				IfNotExists().
				ForeignKey(`("test_section_mapping_id") REFERENCES "test_section_mappings" ("id") ON DELETE CASCADE`).
				Exec(ctx)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewDropTable().
				Model((*testsmodel.ReadingComprehension)(nil)).
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			_, err := db.NewDropTable().
				Model((*testsmodel.SpeakingTopic)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}

