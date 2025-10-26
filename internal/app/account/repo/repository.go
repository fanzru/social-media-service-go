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
	Create(acc *account.Account) error
	GetByID(id int64) (*account.Account, error)
	GetByEmail(email string) (*account.Account, error)
	Update(acc *account.Account) error
	Delete(id int64) error
	SoftDelete(id int64) error
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

// Create creates a new account in the database
func (r *repository) Create(acc *account.Account) error {
	query := `
		INSERT INTO accounts (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	acc.CreatedAt = now
	acc.UpdatedAt = now

	err := r.db.QueryRow(
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
func (r *repository) GetByID(id int64) (*account.Account, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL`

	acc := &account.Account{}
	err := r.db.QueryRow(query, id).Scan(
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
func (r *repository) GetByEmail(email string) (*account.Account, error) {
	query := `
		SELECT id, name, email, password, created_at, updated_at, deleted_at
		FROM accounts
		WHERE email = $1 AND deleted_at IS NULL`

	acc := &account.Account{}
	err := r.db.QueryRow(query, email).Scan(
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
func (r *repository) Update(acc *account.Account) error {
	query := `
		UPDATE accounts
		SET name = $2, email = $3, password = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	acc.UpdatedAt = time.Now()

	result, err := r.db.Exec(
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
func (r *repository) Delete(id int64) error {
	query := `DELETE FROM accounts WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *repository) SoftDelete(id int64) error {
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
