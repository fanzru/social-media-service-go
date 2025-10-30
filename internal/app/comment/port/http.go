package port

import (
	"encoding/json"
	"net/http"

	"github.com/fanzru/social-media-service-go/internal/app/comment"
	"github.com/fanzru/social-media-service-go/internal/app/comment/port/genhttp"
	"github.com/fanzru/social-media-service-go/pkg/middleware"
	"github.com/fanzru/social-media-service-go/pkg/response"
)

// Handler handles HTTP requests for comments
type Handler struct {
	service comment.CommentService
}

// NewHandler creates a new comment handler
func NewHandler(service comment.CommentService) *Handler {
	return &Handler{
		service: service,
	}
}

// PostApiCommentsByPostPostId handles POST /api/comments/by-post/{postId}
func (h *Handler) PostApiCommentsByPostPostId(w http.ResponseWriter, r *http.Request, postId int64) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	var req genhttp.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(r.Context(), "Invalid request body", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	createReq := &comment.CreateCommentRequest{
		Content: req.Content,
		PostID:  postId,
	}

	createdComment, err := h.service.CreateComment(r.Context(), createReq, userID)
	if err != nil {
		if err.Error() == "post not found" {
			response.NotFound(r.Context(), "Post not found", []string{err.Error()}).Send(w, http.StatusNotFound)
			return
		}
		response.InternalServerError(r.Context(), "Failed to create comment", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Comment created successfully", createdComment).Send(w, http.StatusCreated)
}

// GetApiPostsPostIdComments handles GET /api/posts/{postId}/comments
func (h *Handler) GetApiCommentsByPostPostId(w http.ResponseWriter, r *http.Request, postId int64, params genhttp.GetApiCommentsByPostPostIdParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	comments, err := h.service.GetPostComments(r.Context(), postId, cursor, limit)
	if err != nil {
		response.InternalServerError(r.Context(), "Failed to get comments", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Comments retrieved successfully", comments).Send(w, http.StatusOK)
}

// GetApiCommentsId handles GET /api/comments/{id}
func (h *Handler) GetApiCommentsId(w http.ResponseWriter, r *http.Request, id int64) {
	fetchedComment, err := h.service.GetComment(r.Context(), id)
	if err != nil {
		response.NotFound(r.Context(), "Comment not found", []string{err.Error()}).Send(w, http.StatusNotFound)
		return
	}

	response.Success(r.Context(), "Comment retrieved successfully", fetchedComment).Send(w, http.StatusOK)
}

// PutApiCommentsId handles PUT /api/comments/{id}
func (h *Handler) PutApiCommentsId(w http.ResponseWriter, r *http.Request, id int64) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	var req genhttp.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(r.Context(), "Invalid request body", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	updateReq := &comment.UpdateCommentRequest{
		Content: req.Content,
	}

	updatedComment, err := h.service.UpdateComment(r.Context(), id, updateReq, userID)
	if err != nil {
		if err.Error() == "comment not found" {
			response.NotFound(r.Context(), "Comment not found", []string{err.Error()}).Send(w, http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized" {
			response.Forbidden(r.Context(), "Not authorized to update this comment", []string{err.Error()}).Send(w, http.StatusForbidden)
			return
		}
		response.InternalServerError(r.Context(), "Failed to update comment", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Comment updated successfully", updatedComment).Send(w, http.StatusOK)
}

// DeleteApiCommentsId handles DELETE /api/comments/{id}
func (h *Handler) DeleteApiCommentsId(w http.ResponseWriter, r *http.Request, id int64) {
	userID, exists := middleware.GetUserID(r.Context())
	if !exists || userID == 0 {
		response.Unauthorized(r.Context(), "User not authenticated", []string{}).Send(w, http.StatusUnauthorized)
		return
	}

	err := h.service.DeleteComment(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "comment not found" {
			response.NotFound(r.Context(), "Comment not found", []string{err.Error()}).Send(w, http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized" {
			response.Forbidden(r.Context(), "Not authorized to delete this comment", []string{err.Error()}).Send(w, http.StatusForbidden)
			return
		}
		response.InternalServerError(r.Context(), "Failed to delete comment", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "Comment deleted successfully", nil).Send(w, http.StatusOK)
}

// GetApiCommentsUserUserId handles GET /api/comments/user/{userId}
func (h *Handler) GetApiCommentsUserUserId(w http.ResponseWriter, r *http.Request, userId int64, params genhttp.GetApiCommentsUserUserIdParams) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	comments, err := h.service.GetUserComments(r.Context(), userId, cursor, limit)
	if err != nil {
		response.InternalServerError(r.Context(), "Failed to get user comments", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	response.Success(r.Context(), "User comments retrieved successfully", comments).Send(w, http.StatusOK)
}

// Implement the generated interface
var _ genhttp.ServerInterface = (*Handler)(nil)
