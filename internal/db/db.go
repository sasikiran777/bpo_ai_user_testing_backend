package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func ConnectPostgres(databaseURL string) (*bun.DB, error) {

	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(databaseURL),
	))
	sqldb.SetMaxOpenConns(25)                 // Maximum open connections
	sqldb.SetMaxIdleConns(10)                 // Maximum idle connections
	sqldb.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	sqldb.SetConnMaxIdleTime(5 * time.Minute) // Idle connection timeout

	// Test the connection
	if err := sqldb.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
		return nil, err
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	return db, nil
}
