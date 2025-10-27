package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository implements health repository interface
type Repository struct {
	db interface{}
}

// NewRepository creates a new health repository
func NewRepository(db interface{}) *Repository {
	return &Repository{
		db: db,
	}
}

// PingDatabase checks database connectivity
func (r *Repository) PingDatabase(ctx context.Context) error {
	// Type assertion to get the underlying database connection
	switch db := r.db.(type) {
	case *sql.DB:
		return db.PingContext(ctx)
	case interface{ PingContext(context.Context) error }:
		return db.PingContext(ctx)
	default:
		return fmt.Errorf("unsupported database type")
	}
}

// PingRedis checks Redis connectivity (placeholder for future implementation)
func (r *Repository) PingRedis(ctx context.Context) error {
	// TODO: Implement Redis ping when Redis is added
	return nil
}

// PingExternalAPI checks external API connectivity (placeholder for future implementation)
func (r *Repository) PingExternalAPI(ctx context.Context) error {
	// TODO: Implement external API ping when needed
	return nil
}
