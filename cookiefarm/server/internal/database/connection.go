package database

import (
	"context"
	"database/sql"
	_ "embed" // Embed the SQL schema for initializing the database.
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver for database/sql.
)

//go:embed schema.sql
var schemaSQL string

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Tuning del pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, err
	}

	return db, nil
}
