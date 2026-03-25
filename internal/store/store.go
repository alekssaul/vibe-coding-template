package store

import (
	"database/sql"

	"github.com/alekssaul/template/internal/store/db"
	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database connection and sqlc queries.
type Store struct {
	db      *sql.DB
	queries *db.Queries
}

// New opens the SQLite database, runs migrations, and returns a Store.
func New(dbPath string) (*Store, error) {
	sqlDb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// SQLite is single-writer; cap to 1 open connection to avoid locking errors.
	sqlDb.SetMaxOpenConns(1)

	s := &Store{
		db:      sqlDb,
		queries: db.New(sqlDb),
	}
	if err := s.migrate(); err != nil {
		sqlDb.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
