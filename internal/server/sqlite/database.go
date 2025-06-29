// Package sqlite provides some basic functionality for interacting with a SQLite database.
package sqlite

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	_ "github.com/joho/godotenv/autoload"
)

//go:embed schema.sql
var sqlSchema string

var (
	dbPath = config.GetEnv("DB_URL", filepath.Join(config.GetExecutableDir(), "cookiefarm.db"))
	DBPool *sqlitex.Pool
)

const POOL_SIZE = 20

// InitDB initializes the database schema using the SQL schema embedded in the code.
// It runs the SQL schema to set up tables and structures in the database.
func InitDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn := DBPool.Get(ctx)
	if conn == nil {
		return nil
	}
	defer DBPool.Put(conn)

	if err := sqlitex.ExecScript(conn, sqlSchema); err != nil {
		return err
	}

	logger.Log.Info().Msg("Database schema initialized successfully")
	return nil
}

// New initializes the database connection and schema.
// It opens a connection to the SQLite database and initializes the schema by calling InitDB.
// Returns the database connection object.
func New() *sqlitex.Pool {
	uri := fmt.Sprintf("file:%s?_busy_timeout=5000&_foreign_keys=on", dbPath)
	flags := sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE | sqlite.SQLITE_OPEN_URI

	db, err := sqlitex.Open(uri, flags, POOL_SIZE)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to open database")
	}
	DBPool = db
	if err := InitDB(); err != nil {
		logger.Log.Fatal().Err(err).Msg("Database initialization failed")
	}

	return db
}

// Close closes the database connection.
// It disconnects from the database and logs the disconnection message.
func Close() error {
	logger.Log.Info().Str("path", dbPath).Msg("Disconnected from database")
	return DBPool.Close()
}

// FlagsNumber returns the number of flags in the database.
// It queries the database to count the flags and returns the total number.
func FlagsNumber(ctx context.Context) (int, error) {
	conn := DBPool.Get(ctx)
	if conn == nil {
		return 0, errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare("SELECT COUNT(*) FROM flags")
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to prepare COUNT query")
		return 0, err
	}
	defer stmt.Finalize()

	hasRow, err := stmt.Step()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to step through COUNT query")
		return 0, err
	}
	if !hasRow {
		return 0, errors.New("no row returned")
	}

	count := stmt.ColumnInt(0)

	logger.Log.Debug().
		Int("count", count).
		Msg("Flags number")

	return count, nil
}

// BindParams binds parameters to a prepared SQLite statement.
func BindParams(stmt *sqlite.Stmt, args ...any) (*sqlite.Stmt, error) {
	for i, arg := range args {
		idx := i + 1
		switch v := arg.(type) {
		case uint:
			stmt.BindInt64(idx, int64(v))
		case uint16:
			stmt.BindInt64(idx, int64(v))
		case uint32:
			stmt.BindInt64(idx, int64(v))
		case uint64:
			stmt.BindInt64(idx, int64(v))
		case int:
			stmt.BindInt64(idx, int64(v))
		case int64:
			stmt.BindInt64(idx, v)
		case float64:
			stmt.BindFloat(idx, v)
		case string:
			stmt.BindText(idx, v)
		case time.Time:
			stmt.BindText(idx, v.Format(time.RFC3339Nano))
		case nil:
			stmt.BindNull(idx)
		default:
			return nil, fmt.Errorf("unsupported bind type %T at index %d", arg, idx)
		}
	}

	return stmt, nil
}
