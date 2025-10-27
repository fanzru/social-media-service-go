package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
	"github.com/fanzru/social-media-service-go/internal/app/post"
	"github.com/fanzru/social-media-service-go/pkg/sqlwrap"
)

// Repository implements post repository interface
type Repository struct {
	db interface{} // Can be *sql.DB or *sqlwrap.DB
}

// NewRepository creates a new post repository
func NewRepository(db interface{}) *Repository {
	return &Repository{db: db}
}

// Create creates a new post
func (r *Repository) Create(ctx context.Context, post *post.Post) error {
	query := `
		INSERT INTO posts (caption, image_path, image_url, creator_id, creator_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	var err error
	if db, ok := r.db.(*sql.DB); ok {
		err = db.QueryRowContext(ctx, query, post.Caption, post.ImagePath, post.ImageURL, post.CreatorID, post.CreatorName, post.CreatedAt, post.UpdatedAt).Scan(&post.ID)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		err = db.QueryRowContext(ctx, query, post.Caption, post.ImagePath, post.ImageURL, post.CreatorID, post.CreatorName, post.CreatedAt, post.UpdatedAt).Scan(&post.ID)
	}

	return err
}

// GetByID retrieves a post by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*post.Post, error) {
	query := `
		SELECT id, caption, image_path, image_url, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var p post.Post
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		err = db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Caption, &p.ImagePath, &p.ImageURL, &p.CreatorID, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		err = db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Caption, &p.ImagePath, &p.ImageURL, &p.CreatorID, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetByCreatorID retrieves posts by creator ID with cursor-based pagination
func (r *Repository) GetByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*post.PostListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, caption, image_path, image_url, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM posts
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

	var posts []post.Post
	for rows.Next() {
		var p post.Post
		err := rows.Scan(&p.ID, &p.Caption, &p.ImagePath, &p.ImageURL, &p.CreatorID, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &post.PostListResponse{
		Posts:   posts,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}, nil
}

// GetAll retrieves all posts with cursor-based pagination
func (r *Repository) GetAll(ctx context.Context, cursor string, limit int) (*post.PostListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, caption, image_path, image_url, creator_id, creator_name, created_at, updated_at, deleted_at
		FROM posts
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if cursor != "" {
		query += ` AND created_at < $1`
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

	var posts []post.Post
	for rows.Next() {
		var p post.Post
		err := rows.Scan(&p.ID, &p.Caption, &p.ImagePath, &p.ImageURL, &p.CreatorID, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return &post.PostListResponse{
		Posts:   posts,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}, nil
}

// Update updates an existing post
func (r *Repository) Update(ctx context.Context, post *post.Post) error {
	query := `
		UPDATE posts 
		SET caption = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	post.UpdatedAt = time.Now()

	var err error
	if db, ok := r.db.(*sql.DB); ok {
		_, err = db.ExecContext(ctx, query, post.Caption, post.UpdatedAt, post.ID)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		_, err = db.ExecContext(ctx, query, post.Caption, post.UpdatedAt, post.ID)
	}

	return err
}

// SoftDelete soft deletes a post
func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE posts SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()
	var err error
	if db, ok := r.db.(*sql.DB); ok {
		_, err = db.ExecContext(ctx, query, now, id)
	} else if db, ok := r.db.(*sqlwrap.DB); ok {
		_, err = db.ExecContext(ctx, query, now, id)
	}

	return err
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

// GetPostsSortedByComments gets posts sorted by comment count with cursor-based pagination
func (r *Repository) GetPostsSortedByComments(ctx context.Context, cursor string, limit int) (*post.PostListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, caption, image_path, image_url, creator_id, creator_name, created_at, updated_at, deleted_at, comment_count
		FROM posts_with_comment_count
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if cursor != "" {
		query += ` AND (comment_count < $1 OR (comment_count = $1 AND created_at < $2))`
		args = append(args, cursor)
	}

	query += ` ORDER BY comment_count DESC, created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
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

	var posts []post.Post
	for rows.Next() {
		var p post.Post
		err := rows.Scan(&p.ID, &p.Caption, &p.ImagePath, &p.ImageURL, &p.CreatorID, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.CommentCount)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		nextCursor = fmt.Sprintf("%d", posts[len(posts)-1].CommentCount)
	}

	return &post.PostListResponse{
		Posts:   posts,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}, nil
}
