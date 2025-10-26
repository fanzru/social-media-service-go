package sqlwrap

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/logger"
)

// DB wraps sql.DB to add automatic query logging with execution time
type DB struct {
	*sql.DB
	logger *logger.Logger
}

// Tx wraps sql.Tx to add automatic query logging with execution time
type Tx struct {
	*sql.Tx
	logger *logger.Logger
}

// Stmt wraps sql.Stmt to add automatic query logging with execution time
type Stmt struct {
	*sql.Stmt
	logger *logger.Logger
	query  string
}

// NewDB creates a new DB wrapper around sql.DB with logging
func NewDB(db *sql.DB) *DB {
	return &DB{
		DB:     db,
		logger: logger.GetGlobal(),
	}
}

// cleanQuery removes extra whitespace and makes query more readable
func cleanQuery(query string) string {
	// Replace multiple whitespace characters with single space
	re := regexp.MustCompile(`\s+`)
	cleaned := re.ReplaceAllString(query, " ")

	// Trim leading and trailing spaces
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRow(query, args...)
	duration := time.Since(start)

	db.logger.Info("Database QueryRow executed",
		"query", cleanQuery(query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// QueryRowContext executes a query that returns a single row with context
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	db.logger.Info("Database QueryRowContext executed",
		"query", cleanQuery(query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.Query(query, args...)
	duration := time.Since(start)

	if err != nil {
		db.logger.Error("Database Query failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		db.logger.Info("Database Query executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// QueryContext executes a query that returns rows with context
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		db.logger.Error("Database QueryContext failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		db.logger.Info("Database QueryContext executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.Exec(query, args...)
	duration := time.Since(start)

	if err != nil {
		db.logger.Error("Database Exec failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		db.logger.Info("Database Exec executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// ExecContext executes a query without returning any rows with context
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		db.logger.Error("Database ExecContext failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		db.logger.Info("Database ExecContext executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// Prepare creates a prepared statement
func (db *DB) Prepare(query string) (*Stmt, error) {
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:   stmt,
		logger: db.logger,
		query:  query,
	}, nil
}

// PrepareContext creates a prepared statement with context
func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:   stmt,
		logger: db.logger,
		query:  query,
	}, nil
}

// Begin starts a transaction
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:     tx,
		logger: db.logger,
	}, nil
}

// BeginTx starts a transaction with context and options
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:     tx,
		logger: db.logger,
	}, nil
}

// Tx methods

// QueryRow executes a query that returns a single row within transaction
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := tx.Tx.QueryRow(query, args...)
	duration := time.Since(start)

	tx.logger.Info("Database Transaction QueryRow executed",
		"query", cleanQuery(query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// QueryRowContext executes a query that returns a single row within transaction with context
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := tx.Tx.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	tx.logger.Info("Database Transaction QueryRowContext executed",
		"query", cleanQuery(query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// Query executes a query that returns rows within transaction
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.Query(query, args...)
	duration := time.Since(start)

	if err != nil {
		tx.logger.Error("Database Transaction Query failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		tx.logger.Info("Database Transaction Query executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// QueryContext executes a query that returns rows within transaction with context
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		tx.logger.Error("Database Transaction QueryContext failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		tx.logger.Info("Database Transaction QueryContext executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// Exec executes a query without returning any rows within transaction
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.Exec(query, args...)
	duration := time.Since(start)

	if err != nil {
		tx.logger.Error("Database Transaction Exec failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		tx.logger.Info("Database Transaction Exec executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// ExecContext executes a query without returning any rows within transaction with context
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		tx.logger.Error("Database Transaction ExecContext failed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		tx.logger.Info("Database Transaction ExecContext executed",
			"query", cleanQuery(query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// Prepare creates a prepared statement within transaction
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	stmt, err := tx.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:   stmt,
		logger: tx.logger,
		query:  query,
	}, nil
}

// PrepareContext creates a prepared statement within transaction with context
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:   stmt,
		logger: tx.logger,
		query:  query,
	}, nil
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	return tx.Tx.Commit()
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback()
}

// Stmt methods

// QueryRow executes a prepared statement that returns a single row
func (stmt *Stmt) QueryRow(args ...interface{}) *sql.Row {
	start := time.Now()
	row := stmt.Stmt.QueryRow(args...)
	duration := time.Since(start)

	stmt.logger.Info("Database Prepared Statement QueryRow executed",
		"query", cleanQuery(stmt.query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// QueryRowContext executes a prepared statement that returns a single row with context
func (stmt *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	start := time.Now()
	row := stmt.Stmt.QueryRowContext(ctx, args...)
	duration := time.Since(start)

	stmt.logger.Info("Database Prepared Statement QueryRowContext executed",
		"query", cleanQuery(stmt.query),
		"args", args,
		"exec_time_ms", duration.Milliseconds(),
		"exec_time_ns", duration.Nanoseconds(),
	)

	return row
}

// Query executes a prepared statement that returns rows
func (stmt *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := stmt.Stmt.Query(args...)
	duration := time.Since(start)

	if err != nil {
		stmt.logger.Error("Database Prepared Statement Query failed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		stmt.logger.Info("Database Prepared Statement Query executed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// QueryContext executes a prepared statement that returns rows with context
func (stmt *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := stmt.Stmt.QueryContext(ctx, args...)
	duration := time.Since(start)

	if err != nil {
		stmt.logger.Error("Database Prepared Statement QueryContext failed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		stmt.logger.Info("Database Prepared Statement QueryContext executed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return rows, err
}

// Exec executes a prepared statement without returning any rows
func (stmt *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := stmt.Stmt.Exec(args...)
	duration := time.Since(start)

	if err != nil {
		stmt.logger.Error("Database Prepared Statement Exec failed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		stmt.logger.Info("Database Prepared Statement Exec executed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// ExecContext executes a prepared statement without returning any rows with context
func (stmt *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := stmt.Stmt.ExecContext(ctx, args...)
	duration := time.Since(start)

	if err != nil {
		stmt.logger.Error("Database Prepared Statement ExecContext failed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
			"error", err.Error(),
		)
	} else {
		stmt.logger.Info("Database Prepared Statement ExecContext executed",
			"query", cleanQuery(stmt.query),
			"args", args,
			"exec_time_ms", duration.Milliseconds(),
			"exec_time_ns", duration.Nanoseconds(),
		)
	}

	return result, err
}

// Close closes the prepared statement
func (stmt *Stmt) Close() error {
	return stmt.Stmt.Close()
}
