package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"ai_testing/internal/config"
	db "ai_testing/internal/db"
	testsmodel "ai_testing/internal/modules/tests/model"

	"github.com/uptrace/bun"
)

func main() {
	cfg := config.Load()
	if !cfg.RunSeeder {
		fmt.Println("seeder_disabled")
		return
	}

	database, err := db.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer database.Close()

	ctx := context.Background()

	english, err := ensureTest(ctx, database, testsmodel.Test{
		Name:        "English",
		Code:        "english",
		Description: "Measures real-world English communication with timed writing, comprehension, and speaking.",
		Instruction: []string{
			"You can take this test only once.",
			"Do not refresh/close the tab or switch away. Leaving the test may mark it as failed.",
			"Writing: 5 minutes. Reading: 5 minutes. Speaking: 3 minutes (auto-recording)",
			"Microphone access is required for the Speaking section.",
			"Your activity (tab changes / focus loss) is tracked silently during the test.",
		},
		IsActive: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if _, err := ensureTest(ctx, database, testsmodel.Test{
		Name:        "Agentic AI",
		Code:        "agentic_ai",
		Description: "Future module for agentic AI workflows and scenario-based testing.",

		IsActive: false,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	sections := []testsmodel.TestSectionMapping{
		{TestID: english.ID, Name: "Read", Description: "Comprehension and Reading", MaxMarks: 10, MaxTime: 5, IsActive: true},
		{TestID: english.ID, Name: "Write", Description: "Writing", MaxMarks: 10, MaxTime: 5, IsActive: true},
		{TestID: english.ID, Name: "Speak", Description: "Speaking", MaxMarks: 10, MaxTime: 3, IsActive: true},
	}

	for _, s := range sections {
		if err := ensureTestSection(ctx, database, s); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("seed_complete")
}

func ensureTest(ctx context.Context, db *bun.DB, test testsmodel.Test) (*testsmodel.Test, error) {
	var existing testsmodel.Test
	if err := db.NewSelect().
		Model(&existing).
		Where("code = ?", test.Code).
		Limit(1).
		Scan(ctx); err == nil {
		return &existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if _, err := db.NewInsert().
		Model(&test).
		Returning("id").
		Exec(ctx, &test.ID); err != nil {
		return nil, err
	}

	return &test, nil
}

func ensureTestSection(ctx context.Context, db *bun.DB, s testsmodel.TestSectionMapping) error {
	var existing testsmodel.TestSectionMapping
	if err := db.NewSelect().
		Model(&existing).
		Where("test_id = ?", s.TestID).
		Where("name = ?", s.Name).
		Limit(1).
		Scan(ctx); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err := db.NewInsert().
		Model(&s).
		Exec(ctx)
	return err
}
