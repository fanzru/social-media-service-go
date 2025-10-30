package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/account"
	"github.com/fanzru/social-media-service-go/pkg/sqlwrap"
)

// DBInterface defines the database interface that repository needs
type DBInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Repository interface defines the contract for account data operations
type Repository interface {
	Create(ctx context.Context, acc *account.Account) error
	GetByID(ctx context.Context, id int64) (*account.Account, error)
	GetByEmail(ctx context.Context, email string) (*account.Account, error)
	Update(ctx context.Context, acc *account.Account) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, id int64) error
	// ListUserPostImagePaths returns all image_path values for posts created by the user
	ListUserPostImagePaths(ctx context.Context, userID int64) ([]string, error)
	// Transactional helpers
	BeginTx(ctx context.Context) (Tx, error)
	ListUserPostImagePathsTx(ctx context.Context, tx Tx, userID int64) ([]string, error)
	DeleteTx(ctx context.Context, tx Tx, id int64) error
}

// Tx abstracts a SQL transaction used by the repository
type Tx interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// repository implements the Repository interface
type repository struct {
	db DBInterface
}

// NewRepository creates a new account repository
func NewRepository(db interface{}) Repository {
	// Handle both sql.DB and sqlwrap.DB
	switch d := db.(type) {
	case *sql.DB:
		return &repository{db: &sqlDBWrapper{db: d}}
	case *sqlwrap.DB:
		return &repository{db: d}
	default:
		panic("unsupported database type")
	}
}

// sqlDBWrapper wraps sql.DB to implement DBInterface
type sqlDBWrapper struct {
	db *sql.DB
}

func (w *sqlDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.db.QueryRow(query, args...)
}

func (w *sqlDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.Query(query, args...)
}

func (w *sqlDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.db.Exec(query, args...)
}

func (w *sqlDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

// sqlTxWrapper adapts *sql.Tx to our Tx interface
type sqlTxWrapper struct {
	tx *sql.Tx
}

func (t *sqlTxWrapper) Commit() error   { return t.tx.Commit() }
func (t *sqlTxWrapper) Rollback() error { return t.tx.Rollback() }
func (t *sqlTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
func (t *sqlTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// sqlwrapTxAdapter adapts *sqlwrap.Tx to our Tx interface
type sqlwrapTxAdapter struct {
	tx *sqlwrap.Tx
}

func (t *sqlwrapTxAdapter) Commit() error   { return t.tx.Commit() }
func (t *sqlwrapTxAdapter) Rollback() error { return t.tx.Rollback() }
func (t *sqlwrapTxAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
func (t *sqlwrapTxAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// Create creates a new account in the database
func (r *repository) Create(ctx context.Context, acc *account.Account) error {
	query := `
		INSERT INTO accounts (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	acc.CreatedAt = now
	acc.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		acc.Name,
		acc.Email,
		acc.Password,
		acc.CreatedAt,
		acc.UpdatedAt,
	).Scan(&acc.ID)

	return err
}

// GetByID retrieves an account by ID
func (r *repository) GetByID(ctx context.Context, id int64) (*account.Account, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL`

	acc := &account.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&acc.ID,
		&acc.Name,
		&acc.Email,
		&acc.Password,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&acc.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return acc, nil
}

// GetByEmail retrieves an account by email
func (r *repository) GetByEmail(ctx context.Context, email string) (*account.Account, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM accounts
		WHERE email = $1 AND deleted_at IS NULL`

	acc := &account.Account{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&acc.ID,
		&acc.Name,
		&acc.Email,
		&acc.Password,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&acc.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return acc, nil
}

// Update updates an existing account
func (r *repository) Update(ctx context.Context, acc *account.Account) error {
	query := `
		UPDATE accounts
		SET name = $2, email = $3, password = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	acc.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		acc.ID,
		acc.Name,
		acc.Email,
		acc.Password,
		acc.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete permanently deletes an account
func (r *repository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM accounts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SoftDelete soft deletes an account by setting deleted_at
func (r *repository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE accounts
		SET deleted_at = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()

	result, err := r.db.Exec(query, id, now, now)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ListUserPostImagePaths returns all image paths for posts created by the given user
func (r *repository) ListUserPostImagePaths(ctx context.Context, userID int64) ([]string, error) {
	query := `
        SELECT image_path
        FROM posts
        WHERE creator_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imagePaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		if path != "" {
			imagePaths = append(imagePaths, path)
		}
	}

	return imagePaths, nil
}

// BeginTx starts a database transaction
func (r *repository) BeginTx(ctx context.Context) (Tx, error) {
	// Try sqlwrap.DB first
	if db, ok := r.db.(*sqlwrap.DB); ok {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &sqlwrapTxAdapter{tx: tx}, nil
	}
	// Fall back to raw *sql.DB via wrapper
	if wrapper, ok := r.db.(*sqlDBWrapper); ok {
		tx, err := wrapper.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &sqlTxWrapper{tx: tx}, nil
	}
	return nil, sql.ErrConnDone
}

// ListUserPostImagePathsTx returns image paths using a transaction
func (r *repository) ListUserPostImagePathsTx(ctx context.Context, tx Tx, userID int64) ([]string, error) {
	query := `
        SELECT image_path
        FROM posts
        WHERE creator_id = $1 AND deleted_at IS NULL`

	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imagePaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		if path != "" {
			imagePaths = append(imagePaths, path)
		}
	}

	return imagePaths, nil
}

// DeleteTx permanently deletes an account within a transaction
func (r *repository) DeleteTx(ctx context.Context, tx Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	return err
}
