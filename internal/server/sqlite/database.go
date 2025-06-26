// Package sqlite provides some basic functionality for interacting with a SQLite database.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var sqlSchema string

var (
	dbPath = config.GetEnv("DB_URL", filepath.Join(config.GetExecutableDir(), "cookiefarm.db"))
	DB     *sql.DB
)

// InitDB initializes the database schema using the SQL schema embedded in the code.
// It runs the SQL schema to set up tables and structures in the database.
func InitDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := DB.ExecContext(ctx, sqlSchema)
	if err != nil {
		return err
	}

	logger.Log.Info().Msg("Database schema initialized successfully")
	return nil
}

// New initializes the database connection and schema.
// It opens a connection to the SQLite database and initializes the schema by calling InitDB.
// Returns the database connection object.
func New() *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to open database")
	}
	DB = db
	if err := InitDB(); err != nil {
		logger.Log.Fatal().Err(err).Msg("Database initialization failed")
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db
}

// Health checks the health of the database.
// It pings the database to check if it's reachable and returns stats about its connections and usage.
func Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := DB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		logger.Log.Error().Err(err).Msg("Database ping failed")
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := DB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed. Consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime. Consider increasing it."
	}

	return stats
}

// Close closes the database connection.
// It disconnects from the database and logs the disconnection message.
func Close() error {
	logger.Log.Info().Str("path", dbPath).Msg("Disconnected from database")
	return DB.Close()
}

// FlagsNumber returns the number of flags in the database.
// It queries the database to count the flags and returns the total number.
func FlagsNumber(ctx context.Context) (int, error) {
	var count int
	err := DB.
		QueryRowContext(ctx, "SELECT COUNT(*) FROM flags").
		Scan(&count)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to get flags count")
		return 0, err
	}

	logger.Log.Debug().
		Int("count", count).
		Msg("Flags number")
	return count, nil
}
