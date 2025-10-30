package port

import (
	"encoding/json"
	"net/http"

	"github.com/fanzru/social-media-service-go/internal/app/post"
	"github.com/fanzru/social-media-service-go/internal/app/post/port/genhttp"
	"github.com/fanzru/social-media-service-go/pkg/middleware"
	"github.com/fanzru/social-media-service-go/pkg/response"
)

// Handler handles HTTP requests for posts
type Handler struct {
	service post.PostService
}

// NewHandler creates a new post handler
func NewHandler(service post.PostService) *Handler {
	return &Handler{
		service: service,
	}
}

// PostApiPosts handles POST /api/posts
func (h *Handler) PostApiPosts(w http.ResponseWriter, r *http.Request) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.BadRequest(r.Context(), "Failed to parse multipart form", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	caption := r.FormValue("caption")
	if caption == "" {
		response.BadRequest(r.Context(), "Caption is required", []string{"caption field is missing"}).Send(w, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.BadRequest(r.Context(), "Image file is required", []string{"image field is missing"}).Send(w, http.StatusBadRequest)
		return
	}
	defer file.Close()

	createdPost, err := h.service.CreatePostWithImage(r.Context(), userID, caption, file, header)
	if err != nil {
		response.InternalServerError(r.Context(), "Failed to create post", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Post created successfully", createdPost).Send(w, http.StatusCreated)
}

// GetApiPosts handles GET /api/posts
func (h *Handler) GetApiPosts(w http.ResponseWriter, r *http.Request, params genhttp.GetApiPostsParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	posts, err := h.service.GetPostsSortedByComments(r.Context(), cursor, limit)
	if err != nil {
		response.InternalServerError(r.Context(), "Failed to get posts", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Posts retrieved successfully", posts).Send(w, http.StatusOK)
}

// GetApiPostsId handles GET /api/posts/{id}
func (h *Handler) GetApiPostsId(w http.ResponseWriter, r *http.Request, id int64) {
	fetchedPost, err := h.service.GetPostByID(r.Context(), id)
	if err != nil {
		response.NotFound(r.Context(), "Post not found", []string{err.Error()}).Send(w, http.StatusNotFound)
		return
	}

	response.Success(r.Context(), "Post retrieved successfully", fetchedPost).Send(w, http.StatusOK)
}

// PutApiPostsId handles PUT /api/posts/{id}
func (h *Handler) PutApiPostsId(w http.ResponseWriter, r *http.Request, id int64) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	var req genhttp.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(r.Context(), "Invalid request body", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	updateReq := &post.UpdatePostRequest{
		Caption: req.Caption,
	}

	updatedPost, err := h.service.UpdatePost(r.Context(), id, userID, updateReq)
	if err != nil {
		if err.Error() == "post not found" {
			response.NotFound(r.Context(), "Post not found", []string{err.Error()}).Send(w, http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized" {
			response.Forbidden(r.Context(), "Not authorized to update this post", []string{err.Error()}).Send(w, http.StatusForbidden)
			return
		}
		response.InternalServerError(r.Context(), "Failed to update post", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Post updated successfully", updatedPost).Send(w, http.StatusOK)
}

// DeleteApiPostsId handles DELETE /api/posts/{id}
func (h *Handler) DeleteApiPostsId(w http.ResponseWriter, r *http.Request, id int64) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	err := h.service.DeletePost(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "post not found" {
			response.NotFound(r.Context(), "Post not found", []string{err.Error()}).Send(w, http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized" {
			response.Forbidden(r.Context(), "Not authorized to delete this post", []string{err.Error()}).Send(w, http.StatusForbidden)
			return
		}
		response.InternalServerError(r.Context(), "Failed to delete post", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Post deleted successfully", nil).Send(w, http.StatusOK)
}

// GetApiPostsUserUserId handles GET /api/posts/user/{userId}
func (h *Handler) GetApiPostsByUserUserId(w http.ResponseWriter, r *http.Request, userId int64, params genhttp.GetApiPostsByUserUserIdParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	posts, err := h.service.GetPostsByCreatorID(r.Context(), userId, cursor, limit)
	if err != nil {
		response.InternalServerError(r.Context(), "Failed to get user posts", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "User posts retrieved successfully", posts).Send(w, http.StatusOK)
}

// Implement the generated interface
var _ genhttp.ServerInterface = (*Handler)(nil)
