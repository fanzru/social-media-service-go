package post

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
)

// Post represents a social media post
type Post struct {
	ID          int64      `json:"id" db:"id"`
	Caption     string     `json:"caption" db:"caption"`
	ImagePath   string     `json:"image_path" db:"image_path"`
	ImageURL    string     `json:"image_url" db:"image_url"`
	CreatorID   int64      `json:"creator_id" db:"creator_id"`
	CreatorName string     `json:"creator_name" db:"creator_name"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Computed fields
	CommentCount int64             `json:"comment_count,omitempty" db:"comment_count"`
	Comments     []comment.Comment `json:"comments,omitempty" db:"comments"`
}

// CreatePostRequest represents the request payload for creating a post
type CreatePostRequest struct {
	Caption string `json:"caption" validate:"required,max=1000"`
	// Image will be handled separately via multipart form
}

// UpdatePostRequest represents the request payload for updating a post
type UpdatePostRequest struct {
	Caption string `json:"caption" validate:"max=1000"`
}

// PostListRequest represents the request payload for listing posts
type PostListRequest struct {
	Cursor string `json:"cursor,omitempty"` // For cursor-based pagination
	Limit  int    `json:"limit,omitempty"`  // Default to 20, max 100
}

// PostListResponse represents the response payload for listing posts
type PostListResponse struct {
	Posts   []Post `json:"posts"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
}

// PostResponse represents the response payload for a single post
type PostResponse struct {
	Post Post `json:"post"`
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id int64) (*Post, error)
	GetByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*PostListResponse, error)
	GetAll(ctx context.Context, cursor string, limit int) (*PostListResponse, error)
	Update(ctx context.Context, post *Post) error
	SoftDelete(ctx context.Context, id int64) error
	GetCommentCount(ctx context.Context, postID int64) (int64, error)
	GetLastComments(ctx context.Context, postID int64, limit int) ([]comment.Comment, error)
	GetPostsSortedByComments(ctx context.Context, cursor string, limit int) (*PostListResponse, error)
}

// PostService defines the interface for post business logic
type PostService interface {
	CreatePost(ctx context.Context, req *CreatePostRequest, creatorID int64, imagePath string) (*Post, error)
	CreatePostWithImage(ctx context.Context, creatorID int64, caption string, file multipart.File, header *multipart.FileHeader) (*Post, error)
	GetPost(ctx context.Context, id int64) (*Post, error)
	GetPostByID(ctx context.Context, id int64) (*Post, error)
	GetUserPosts(ctx context.Context, creatorID int64, cursor string, limit int) (*PostListResponse, error)
	GetPostsByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*PostListResponse, error)
	GetAllPosts(ctx context.Context, cursor string, limit int) (*PostListResponse, error)
	GetPostsSortedByComments(ctx context.Context, cursor string, limit int) (*PostListResponse, error)
	UpdatePost(ctx context.Context, id int64, creatorID int64, req *UpdatePostRequest) (*Post, error)
	DeletePost(ctx context.Context, id int64, creatorID int64) error
	GetPostsWithComments(ctx context.Context, cursor string, limit int) (*PostListResponse, error)
}
