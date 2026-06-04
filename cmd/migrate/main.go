package main

import (
	"context"
	"fmt"
	"os"

	"github.com/uptrace/bun/migrate"

	"ai_testing/internal/config"
	db "ai_testing/internal/db"
	"ai_testing/migrations"
)

func main() {
	cmd := "up"
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}

	cfg := config.Load()
	database, err := db.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer database.Close()

	ctx := context.Background()
	migrator := migrate.NewMigrator(database, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch cmd {
	case "up":
		if _, err := migrator.Migrate(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "down":
		if _, err := migrator.Rollback(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "status":
		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, m := range ms {
			var s any = nil
			if m.IsApplied() {
				s = m.MigratedAt
			}
			fmt.Printf("%s\t%v\n", m.Name, s)
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate [up|down|status]")
		os.Exit(2)
	}
}
