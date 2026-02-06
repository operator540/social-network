package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// createCommentRequest — тело запроса создания комментария
type createCommentRequest struct {
	Content string `json:"content"`
}

// createComment обрабатывает POST /v1/posts/{id}/comments
func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID поста")
		return
	}

	var req createCommentRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Content == "" {
		jsonError(w, http.StatusBadRequest, "текст комментария обязателен")
		return
	}

	userID := getUserID(r)
	comment, err := h.commentService.Create(postID, userID, req.Content)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка создания комментария")
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

// getComments обрабатывает GET /v1/posts/{id}/comments
func (h *Handler) getComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID поста")
		return
	}

	comments, err := h.commentService.GetByPostID(postID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения комментариев")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}
