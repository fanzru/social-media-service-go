package app

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
	"github.com/fanzru/social-media-service-go/internal/app/post"
	"github.com/fanzru/social-media-service-go/pkg/storage"
)

// Service implements post service interface
type Service struct {
	repo         post.PostRepository
	commentRepo  comment.CommentRepository
	imageStorage *storage.ImageStorageService
}

// NewService creates a new post service
func NewService(repo post.PostRepository, commentRepo comment.CommentRepository, imageStorage *storage.ImageStorageService) *Service {
	return &Service{
		repo:         repo,
		commentRepo:  commentRepo,
		imageStorage: imageStorage,
	}
}

// CreatePostWithImage creates a new post with image upload (HTTP handler version)
func (s *Service) CreatePostWithImage(ctx context.Context, creatorID int64, caption string, file multipart.File, header *multipart.FileHeader) (*post.Post, error) {
	req := &post.CreatePostRequest{
		Caption: caption,
	}
	return s.createPostWithImage(ctx, req, creatorID, file, header)
}

// createPostWithImage creates a new post with image upload (internal method)
func (s *Service) createPostWithImage(ctx context.Context, req *post.CreatePostRequest, creatorID int64, file multipart.File, header *multipart.FileHeader) (*post.Post, error) {
	// Validate caption
	if err := s.validateCaption(req.Caption); err != nil {
		return nil, fmt.Errorf("invalid caption: %w", err)
	}

	// Process and upload image
	imagePath, imageURL, err := s.imageStorage.ProcessAndUploadImage(file, header)
	if err != nil {
		return nil, fmt.Errorf("failed to process and upload image: %w", err)
	}

	// Create post
	newPost := &post.Post{
		Caption:     req.Caption,
		ImagePath:   imagePath,
		ImageURL:    imageURL,
		CreatorID:   creatorID,
		CreatorName: "", // Will be populated from account service
	}

	if err := s.repo.Create(ctx, newPost); err != nil {
		// If post creation fails, try to delete the uploaded image
		s.imageStorage.DeleteImage(imagePath)
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return newPost, nil
}

// CreatePost creates a new post (legacy method for backward compatibility)
func (s *Service) CreatePost(ctx context.Context, req *post.CreatePostRequest, creatorID int64, imagePath string) (*post.Post, error) {
	// Validate caption
	if err := s.validateCaption(req.Caption); err != nil {
		return nil, fmt.Errorf("invalid caption: %w", err)
	}

	// Generate image URL from path
	imageURL := s.generateImageURL(imagePath)

	// Create post
	newPost := &post.Post{
		Caption:     req.Caption,
		ImagePath:   imagePath,
		ImageURL:    imageURL,
		CreatorID:   creatorID,
		CreatorName: "", // Will be populated from account service
	}

	if err := s.repo.Create(ctx, newPost); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return newPost, nil
}

// GetPost retrieves a post by ID
func (s *Service) GetPost(ctx context.Context, id int64) (*post.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Get comment count
	commentCount, err := s.repo.GetCommentCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment count: %w", err)
	}
	post.CommentCount = commentCount

	// Get last 2 comments
	comments, err := s.repo.GetLastComments(ctx, id, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to get last comments: %w", err)
	}
	post.Comments = comments

	return post, nil
}

// GetPostByID is an alias for GetPost for backward compatibility
func (s *Service) GetPostByID(ctx context.Context, id int64) (*post.Post, error) {
	return s.GetPost(ctx, id)
}

// GetUserPosts retrieves posts by creator ID
func (s *Service) GetUserPosts(ctx context.Context, creatorID int64, cursor string, limit int) (*post.PostListResponse, error) {
	response, err := s.repo.GetByCreatorID(ctx, creatorID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user posts: %w", err)
	}

	// Add comment counts and last comments for each post
	for i := range response.Posts {
		commentCount, err := s.repo.GetCommentCount(ctx, response.Posts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get comment count for post %d: %w", response.Posts[i].ID, err)
		}
		response.Posts[i].CommentCount = commentCount

		comments, err := s.repo.GetLastComments(ctx, response.Posts[i].ID, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to get last comments for post %d: %w", response.Posts[i].ID, err)
		}
		response.Posts[i].Comments = comments
	}

	return response, nil
}

// GetPostsByCreatorID is an alias for GetUserPosts for backward compatibility
func (s *Service) GetPostsByCreatorID(ctx context.Context, creatorID int64, cursor string, limit int) (*post.PostListResponse, error) {
	return s.GetUserPosts(ctx, creatorID, cursor, limit)
}

// GetAllPosts retrieves all posts
func (s *Service) GetAllPosts(ctx context.Context, cursor string, limit int) (*post.PostListResponse, error) {
	response, err := s.repo.GetAll(ctx, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all posts: %w", err)
	}

	// Add comment counts and last comments for each post
	for i := range response.Posts {
		commentCount, err := s.repo.GetCommentCount(ctx, response.Posts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get comment count for post %d: %w", response.Posts[i].ID, err)
		}
		response.Posts[i].CommentCount = commentCount

		comments, err := s.repo.GetLastComments(ctx, response.Posts[i].ID, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to get last comments for post %d: %w", response.Posts[i].ID, err)
		}
		response.Posts[i].Comments = comments
	}

	return response, nil
}

// UpdatePost updates an existing post
func (s *Service) UpdatePost(ctx context.Context, id int64, creatorID int64, req *post.UpdatePostRequest) (*post.Post, error) {
	// Get existing post
	existingPost, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Check if user owns the post
	if existingPost.CreatorID != creatorID {
		return nil, fmt.Errorf("unauthorized: you can only update your own posts")
	}

	// Validate caption
	if err := s.validateCaption(req.Caption); err != nil {
		return nil, fmt.Errorf("invalid caption: %w", err)
	}

	// Update post
	existingPost.Caption = req.Caption
	if err := s.repo.Update(ctx, existingPost); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return existingPost, nil
}

// DeletePost deletes a post
func (s *Service) DeletePost(ctx context.Context, id int64, creatorID int64) error {
	// Get existing post
	existingPost, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	// Check if user owns the post
	if existingPost.CreatorID != creatorID {
		return fmt.Errorf("unauthorized: you can only delete your own posts")
	}

	// Soft delete post
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Delete associated image from storage
	if err := s.imageStorage.DeleteImage(existingPost.ImagePath); err != nil {
		// Log error but don't fail the post deletion
		// Image cleanup can be handled by a background job
		fmt.Printf("Warning: failed to delete image %s: %v\n", existingPost.ImagePath, err)
	}

	return nil
}

// GetPostsWithComments retrieves posts sorted by comment count with last 2 comments
func (s *Service) GetPostsWithComments(ctx context.Context, cursor string, limit int) (*post.PostListResponse, error) {
	response, err := s.repo.GetPostsSortedByComments(ctx, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts sorted by comments: %w", err)
	}

	// Add last 2 comments for each post
	for i := range response.Posts {
		comments, err := s.repo.GetLastComments(ctx, response.Posts[i].ID, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to get last comments for post %d: %w", response.Posts[i].ID, err)
		}
		response.Posts[i].Comments = comments
	}

	return response, nil
}

// GetPostsSortedByComments is an alias for GetPostsWithComments for backward compatibility
func (s *Service) GetPostsSortedByComments(ctx context.Context, cursor string, limit int) (*post.PostListResponse, error) {
	return s.GetPostsWithComments(ctx, cursor, limit)
}

// validateCaption validates the post caption
func (s *Service) validateCaption(caption string) error {
	if len(caption) > 1000 {
		return fmt.Errorf("caption must be at most 1000 characters")
	}
	return nil
}

// generateImageURL generates the public URL for an image
func (s *Service) generateImageURL(imagePath string) string {
	// Extract filename from path
	filename := filepath.Base(imagePath)

	// Convert to JPG format (as per requirement)
	ext := filepath.Ext(filename)
	if ext != ".jpg" && ext != ".jpeg" {
		filename = strings.TrimSuffix(filename, ext) + ".jpg"
	}

	// Use image storage service to generate URL
	// This will handle both S3 and local storage URLs
	return s.imageStorage.GenerateImageURL(filename)
}
