package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
	"github.com/fanzru/social-media-service-go/pkg/sqlwrap"
)

// Repository implements comment repository interface
type Repository struct {
	db interface{} // Can be *sql.DB or *sqlwrap.DB
}

// NewRepository creates a new comment repository
func NewRepository(db interface{}) *Repository {
	return &Repository{db: db}
}

// Create creates a new comment
func (r *Repository) Create(ctx context.Context, comment *comment.Comment) error {
	query := `
		INSERT INTO comments (content, post_id, creator_id, creator_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	var err error
	if db, ok := r.db.(*sql.DB); ok {
		err = db.QueryRowContext(ctx, query, comment.Content, comment.PostID, comment.CreatorID, comment.CreatorName, comment.CreatedAt, comment.UpdatedAt).Scan(&comment.ID)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		err = db.QueryRowContext(ctx, query, comment.Content, comment.PostID, comment.CreatorID, comment.CreatorName, comment.CreatedAt, comment.UpdatedAt).Scan(&comment.ID)
	}

	return err
}

// GetByID retrieves a comment by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*comment.Comment, error) {
	query := `
		SELECT id, content, post_id, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var c comment.Comment
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		err = db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Content, &c.PostID, &c.CreatorID, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		err = db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Content, &c.PostID, &c.CreatorID, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetByPostID retrieves comments by post ID with cursor-based pagination
func (r *Repository) GetByPostID(ctx context.Context, postID int64, cursor string, limit int) (*comment.CommentListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, content, post_id, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM comments
		WHERE post_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{postID}

	if cursor != "" {
		query += ` AND created_at < $2`
		args = append(args, cursor)
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit+1) // Get one extra to check if there are more

	var rows *sql.Rows
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		rows, err = db.QueryContext(ctx, query, args...)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		rows, err = db.QueryContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []comment.Comment
	for rows.Next() {
		var c comment.Comment
		err := rows.Scan(&c.ID, &c.Content, &c.PostID, &c.CreatorID, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	var nextCursor string
	if hasMore && len(comments) > 0 {
		nextCursor = comments[len(comments)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &comment.CommentListResponse{
		Comments: comments,
		Cursor:   nextCursor,
		HasMore:  hasMore,
	}, nil
}

// GetByCreatorID retrieves comments by creator ID with cursor-based pagination
func (r *Repository) GetByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*comment.CommentListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, content, post_id, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM comments
		WHERE creator_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{creatorID}

	if cursor != "" {
		query += ` AND created_at < $2`
		args = append(args, cursor)
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit+1) // Get one extra to check if there are more

	var rows *sql.Rows
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		rows, err = db.QueryContext(ctx, query, args...)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		rows, err = db.QueryContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []comment.Comment
	for rows.Next() {
		var c comment.Comment
		err := rows.Scan(&c.ID, &c.Content, &c.PostID, &c.CreatorID, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	var nextCursor string
	if hasMore && len(comments) > 0 {
		nextCursor = comments[len(comments)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &comment.CommentListResponse{
		Comments: comments,
		Cursor:   nextCursor,
		HasMore:  hasMore,
	}, nil
}

// Update updates an existing comment
func (r *Repository) Update(ctx context.Context, comment *comment.Comment) error {
	query := `
		UPDATE comments 
		SET content = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	comment.UpdatedAt = time.Now()

	var err error
	if db, ok := r.db.(*sql.DB); ok {
		_, err = db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		_, err = db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	}

	return err
}

// SoftDelete soft deletes a comment
func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE comments SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		_, err = db.ExecContext(ctx, query, now, id)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		_, err = db.ExecContext(ctx, query, now, id)
	}

	return err
}

// GetLastComments gets the last N comments for a post
func (r *Repository) GetLastComments(ctx context.Context, postID int64, limit int) ([]comment.Comment, error) {
	if limit <= 0 {
		limit = 2 // Default to 2 as per requirement
	}

	query := `
		SELECT id, content, post_id, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM comments
		WHERE post_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2
	`

	var rows *sql.Rows
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		rows, err = db.QueryContext(ctx, query, postID, limit)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		rows, err = db.QueryContext(ctx, query, postID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []comment.Comment
	for rows.Next() {
		var c comment.Comment
		err := rows.Scan(&c.ID, &c.Content, &c.PostID, &c.CreatorID, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// GetCommentCount gets the comment count for a post
func (r *Repository) GetCommentCount(ctx context.Context, postID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1 AND deleted_at IS NULL`

	var count int64
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		err = db.QueryRowContext(ctx, query, postID).Scan(&count)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		err = db.QueryRowContext(ctx, query, postID).Scan(&count)
	}

	return count, err
}
