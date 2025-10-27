package comment

import (
	"context"
	"time"
)

// Comment represents a comment on a post
type Comment struct {
	ID          int64      `json:"id" db:"id"`
	Content     string     `json:"content" db:"content"`
	PostID      int64      `json:"post_id" db:"post_id"`
	CreatorID   int64      `json:"creator_id" db:"creator_id"`
	CreatorName string     `json:"creator_name" db:"creator_name"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateCommentRequest represents the request payload for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,max=500"`
	PostID  int64  `json:"post_id" validate:"required"`
}

// UpdateCommentRequest represents the request payload for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,max=500"`
}

// CommentListRequest represents the request payload for listing comments
type CommentListRequest struct {
	PostID int64  `json:"post_id" validate:"required"`
	Cursor string `json:"cursor,omitempty"` // For cursor-based pagination
	Limit  int    `json:"limit,omitempty"`  // Default to 20, max 100
}

// CommentListResponse represents the response payload for listing comments
type CommentListResponse struct {
	Comments []Comment `json:"comments"`
	Cursor   string    `json:"cursor,omitempty"`
	HasMore  bool      `json:"has_more"`
}

// CommentResponse represents the response payload for a single comment
type CommentResponse struct {
	Comment Comment `json:"comment"`
}

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id int64) (*Comment, error)
	GetByPostID(ctx context.Context, postID int64, cursor string, limit int) (*CommentListResponse, error)
	GetByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*CommentListResponse, error)
	Update(ctx context.Context, comment *Comment) error
	SoftDelete(ctx context.Context, id int64) error
	GetLastComments(ctx context.Context, postID int64, limit int) ([]Comment, error)
	GetCommentCount(ctx context.Context, postID int64) (int64, error)
}

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(ctx context.Context, req *CreateCommentRequest, creatorID int64) (*Comment, error)
	GetComment(ctx context.Context, id int64) (*Comment, error)
	GetPostComments(ctx context.Context, postID int64, cursor string, limit int) (*CommentListResponse, error)
	GetUserComments(ctx context.Context, creatorID int64, cursor string, limit int) (*CommentListResponse, error)
	UpdateComment(ctx context.Context, id int64, req *UpdateCommentRequest, creatorID int64) (*Comment, error)
	DeleteComment(ctx context.Context, id int64, creatorID int64) error
	GetLastComments(ctx context.Context, postID int64, limit int) ([]Comment, error)
}
