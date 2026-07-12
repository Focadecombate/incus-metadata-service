package db

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

func ConnectDB(cfg *config.Config) (*Queries, error) {
	ctx := context.Background()

	dsn := cfg.Database.DBSource
	if cfg.Database.DBDriver == "sqlite" {
		// Enable WAL journaling and a busy timeout so concurrent cron writes and
		// HTTP reads don't return SQLITE_BUSY. modernc.org/sqlite applies these
		// pragmas per-connection when passed as DSN query params.
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
	}

	db, err := sql.Open(cfg.Database.DBDriver, dsn)
	if err != nil {
		return nil, err
	}

	if cfg.Database.DBDriver == "sqlite" {
		// SQLite allows a single writer; serialize access through one connection
		// to guarantee no SQLITE_BUSY under concurrent access.
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}

	// Check if the database connection is established
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Prepare the queries
	queries := New(db)

	return queries, nil
}
