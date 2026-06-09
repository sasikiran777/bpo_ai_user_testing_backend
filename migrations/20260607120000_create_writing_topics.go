package migrations

import (
	"context"

	testsmodel "ai_testing/internal/modules/tests/model"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*testsmodel.WritingTopic)(nil)).
				IfNotExists().
				ForeignKey(`("test_section_mapping_id") REFERENCES "test_section_mappings" ("id") ON DELETE CASCADE`).
				Exec(ctx)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*testsmodel.WritingTopic)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}
