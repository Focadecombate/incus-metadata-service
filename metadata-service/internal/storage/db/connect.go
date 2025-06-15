package db

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

func ConnectDB(cfg *config.Config) (*Queries, error) {
	ctx := context.Background()
	db, err := sql.Open(cfg.Database.DBDriver, cfg.Database.DBSource)
	if err != nil {
		return nil, err
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
