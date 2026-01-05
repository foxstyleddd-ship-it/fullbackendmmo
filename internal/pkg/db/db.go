package db

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Connect establishes a connection to PostgreSQL
func Connect(connStr string, maxOpenConns, maxIdleConns int) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Info().
		Int("max_open_conns", maxOpenConns).
		Int("max_idle_conns", maxIdleConns).
		Msg("Database connected successfully")

	return db, nil
}

// Close closes the database connection
func Close(db *sqlx.DB) error {
	if db != nil {
		log.Info().Msg("Closing database connection")
		return db.Close()
	}
	return nil
}

// Ping pings the database
func Ping(db *sqlx.DB) error {
	return db.Ping()
}

// InTransaction executes a function within a transaction
func InTransaction(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error().Err(rbErr).Msg("Failed to rollback transaction")
		}
		return err
	}

	return tx.Commit()
}

// NullString converts string to sql.NullString
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// NullInt64 converts int64 to sql.NullInt64
func NullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
