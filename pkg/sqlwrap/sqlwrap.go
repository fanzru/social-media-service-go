package sqlwrap

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/influxdb"
	"github.com/fanzru/social-media-service-go/pkg/logger"
)

// DB wraps sql.DB to add automatic query logging with execution time and metrics
type DB struct {
	*sql.DB
	logger       *logger.Logger
	influxClient *influxdb.Client
}

// Tx wraps sql.Tx to add automatic query logging with execution time and metrics
type Tx struct {
	*sql.Tx
	logger       *logger.Logger
	influxClient *influxdb.Client
}

// Stmt wraps sql.Stmt to add automatic query logging with execution time and metrics
type Stmt struct {
	*sql.Stmt
	logger       *logger.Logger
	influxClient *influxdb.Client
	query        string
}

// NewDB creates a new DB wrapper around sql.DB with logging
func NewDB(db *sql.DB) *DB {
	return &DB{
		DB:     db,
		logger: logger.GetGlobal(),
	}
}

// NewDBWithInfluxDB creates a new DB wrapper around sql.DB with logging and InfluxDB metrics
func NewDBWithInfluxDB(db *sql.DB, influxClient *influxdb.Client) *DB {
	return &DB{
		DB:           db,
		logger:       logger.GetGlobal(),
		influxClient: influxClient,
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

// getOperationFromQuery extracts the operation type from SQL query
func getOperationFromQuery(query string) string {
	query = strings.ToUpper(strings.TrimSpace(query))

	switch {
	case strings.HasPrefix(query, "SELECT"):
		return "SELECT"
	case strings.HasPrefix(query, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(query, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(query, "DELETE"):
		return "DELETE"
	case strings.HasPrefix(query, "CREATE"):
		return "CREATE"
	case strings.HasPrefix(query, "DROP"):
		return "DROP"
	case strings.HasPrefix(query, "ALTER"):
		return "ALTER"
	default:
		return "UNKNOWN"
	}
}

// getTableFromQuery extracts the table name from SQL query
func getTableFromQuery(query string) string {
	query = strings.ToUpper(strings.TrimSpace(query))

	// Simple table extraction - this could be more sophisticated
	words := strings.Fields(query)

	for i, word := range words {
		if word == "FROM" && i+1 < len(words) {
			return strings.ToLower(words[i+1])
		}
		if word == "INTO" && i+1 < len(words) {
			return strings.ToLower(words[i+1])
		}
		if word == "UPDATE" && i+1 < len(words) {
			return strings.ToLower(words[i+1])
		}
	}

	return "unknown"
}

// recordMetrics records database metrics if InfluxDB client is available
func (db *DB) recordMetrics(operation, table string, duration time.Duration, err error) {
	if db.influxClient == nil {
		return
	}

	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}

	tags := map[string]string{
		"group":     "DATABASE",
		"entity":    fmt.Sprintf("%s %s", operation, table),
		"operation": operation,
		"table":     table,
		"code":      status,
	}

	// Record query count
	_ = db.influxClient.WriteCounter("db_queries_total", tags, 1)

	// Record query duration
	_ = db.influxClient.WriteTiming("db_query_duration_ms", tags, duration)
}

// recordTxMetrics records transaction metrics if InfluxDB client is available
func (tx *Tx) recordTxMetrics(operation, table string, duration time.Duration, err error) {
	if tx.influxClient == nil {
		return
	}

	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}

	tags := map[string]string{
		"group":     "DATABASE",
		"entity":    fmt.Sprintf("%s %s", operation, table),
		"operation": operation,
		"table":     table,
		"code":      status,
	}

	// Record query count
	_ = tx.influxClient.WriteCounter("db_queries_total", tags, 1)

	// Record query duration
	_ = tx.influxClient.WriteTiming("db_query_duration_ms", tags, duration)
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	operation := getOperationFromQuery(query)
	table := getTableFromQuery(query)

	row := db.DB.QueryRow(query, args...)
	duration := time.Since(start)

	// Record metrics
	db.recordMetrics(operation, table, duration, nil)

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
	operation := getOperationFromQuery(query)
	table := getTableFromQuery(query)

	rows, err := db.DB.Query(query, args...)
	duration := time.Since(start)

	// Record metrics
	db.recordMetrics(operation, table, duration, err)

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
	operation := getOperationFromQuery(query)
	table := getTableFromQuery(query)

	result, err := db.DB.Exec(query, args...)
	duration := time.Since(start)

	// Record metrics
	db.recordMetrics(operation, table, duration, err)

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
		Stmt:         stmt,
		logger:       db.logger,
		influxClient: db.influxClient,
		query:        query,
	}, nil
}

// PrepareContext creates a prepared statement with context
func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:         stmt,
		logger:       db.logger,
		influxClient: db.influxClient,
		query:        query,
	}, nil
}

// Begin starts a transaction
func (db *DB) Begin() (*Tx, error) {
	start := time.Now()

	tx, err := db.DB.Begin()
	duration := time.Since(start)

	// Record transaction metrics
	if db.influxClient != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILED"
		}

		tags := map[string]string{
			"group":      "DB",
			"entity":     "BEGIN",
			"error_code": status,
		}

		_ = db.influxClient.WriteTiming("db_transaction_duration_ms", tags, duration)
	}

	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:           tx,
		logger:       db.logger,
		influxClient: db.influxClient,
	}, nil
}

// BeginTx starts a transaction with context and options
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	start := time.Now()

	tx, err := db.DB.BeginTx(ctx, opts)
	duration := time.Since(start)

	// Record transaction metrics
	if db.influxClient != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILED"
		}

		tags := map[string]string{
			"group":      "DB",
			"entity":     "BEGIN_TX",
			"error_code": status,
		}

		_ = db.influxClient.WriteTiming("db_transaction_duration_ms", tags, duration)
	}

	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:           tx,
		logger:       db.logger,
		influxClient: db.influxClient,
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
		Stmt:         stmt,
		logger:       tx.logger,
		influxClient: tx.influxClient,
		query:        query,
	}, nil
}

// PrepareContext creates a prepared statement within transaction with context
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &Stmt{
		Stmt:         stmt,
		logger:       tx.logger,
		influxClient: tx.influxClient,
		query:        query,
	}, nil
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	start := time.Now()

	err := tx.Tx.Commit()
	duration := time.Since(start)

	// Record transaction metrics
	if tx.influxClient != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILED"
		}

		tags := map[string]string{
			"group":      "DB",
			"entity":     "COMMIT",
			"error_code": status,
		}

		_ = tx.influxClient.WriteTiming("db_transaction_duration_ms", tags, duration)
	}

	return err
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	start := time.Now()

	err := tx.Tx.Rollback()
	duration := time.Since(start)

	// Record transaction metrics
	if tx.influxClient != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILED"
		}

		tags := map[string]string{
			"group":      "DB",
			"entity":     "ROLLBACK",
			"error_code": status,
		}

		_ = tx.influxClient.WriteTiming("db_transaction_duration_ms", tags, duration)
	}

	return err
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
