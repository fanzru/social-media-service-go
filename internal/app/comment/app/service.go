package app

import (
	"context"
	"fmt"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
	"github.com/fanzru/social-media-service-go/internal/app/post"
)

// Service implements comment service interface
type Service struct {
	repo     comment.CommentRepository
	postRepo post.PostRepository
}

// NewService creates a new comment service
func NewService(repo comment.CommentRepository, postRepo post.PostRepository) *Service {
	return &Service{
		repo:     repo,
		postRepo: postRepo,
	}
}

// CreateComment creates a new comment
func (s *Service) CreateComment(ctx context.Context, req *comment.CreateCommentRequest, creatorID int64) (*comment.Comment, error) {
	// Validate content
	if err := s.validateContent(req.Content); err != nil {
		return nil, fmt.Errorf("invalid content: %w", err)
	}

	// Check if post exists
	_, err := s.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Create comment
	newComment := &comment.Comment{
		Content:     req.Content,
		PostID:      req.PostID,
		CreatorID:   creatorID,
		CreatorName: "", // Will be populated from account service
	}

	if err := s.repo.Create(ctx, newComment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return newComment, nil
}

// GetComment retrieves a comment by ID
func (s *Service) GetComment(ctx context.Context, id int64) (*comment.Comment, error) {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

// GetPostComments retrieves comments for a specific post
func (s *Service) GetPostComments(ctx context.Context, postID int64, cursor string, limit int) (*comment.CommentListResponse, error) {
	// Check if post exists
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	response, err := s.repo.GetByPostID(ctx, postID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get post comments: %w", err)
	}

	return response, nil
}

// GetUserComments retrieves comments by creator ID
func (s *Service) GetUserComments(ctx context.Context, creatorID int64, cursor string, limit int) (*comment.CommentListResponse, error) {
	response, err := s.repo.GetByCreatorID(ctx, creatorID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user comments: %w", err)
	}

	return response, nil
}

// UpdateComment updates an existing comment
func (s *Service) UpdateComment(ctx context.Context, id int64, req *comment.UpdateCommentRequest, creatorID int64) (*comment.Comment, error) {
	// Get existing comment
	existingComment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Check if user owns the comment
	if existingComment.CreatorID != creatorID {
		return nil, fmt.Errorf("unauthorized: you can only update your own comments")
	}

	// Validate content
	if err := s.validateContent(req.Content); err != nil {
		return nil, fmt.Errorf("invalid content: %w", err)
	}

	// Update comment
	existingComment.Content = req.Content
	if err := s.repo.Update(ctx, existingComment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return existingComment, nil
}

// DeleteComment deletes a comment
func (s *Service) DeleteComment(ctx context.Context, id int64, creatorID int64) error {
	// Get existing comment
	existingComment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Check if user owns the comment
	if existingComment.CreatorID != creatorID {
		return fmt.Errorf("unauthorized: you can only delete your own comments")
	}

	// Soft delete comment
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// GetLastComments gets the last N comments for a post
func (s *Service) GetLastComments(ctx context.Context, postID int64, limit int) ([]comment.Comment, error) {
	comments, err := s.repo.GetLastComments(ctx, postID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get last comments: %w", err)
	}

	return comments, nil
}

// validateContent validates the comment content
func (s *Service) validateContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("content is required")
	}
	if len(content) > 500 {
		return fmt.Errorf("content must be at most 500 characters")
	}
	return nil
}
