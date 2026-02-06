package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"social-network/internal/service"
)

// createPostRequest — тело запроса создания поста
type createPostRequest struct {
	Content string `json:"content"`
}

// createPost обрабатывает POST /v1/posts
func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	var req createPostRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Content == "" {
		jsonError(w, http.StatusBadRequest, "текст поста обязателен")
		return
	}

	userID := getUserID(r)
	post, err := h.postService.Create(userID, req.Content)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка создания поста")
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

// deletePost обрабатывает DELETE /v1/posts/{id}
func (h *Handler) deletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID поста")
		return
	}

	userID := getUserID(r)
	if err := h.postService.Delete(postID, userID); err != nil {
		if err == service.ErrNotPostOwner {
			jsonError(w, http.StatusForbidden, "вы не являетесь автором поста")
			return
		}
		if err == service.ErrPostNotFound {
			jsonError(w, http.StatusNotFound, "пост не найден")
			return
		}
		jsonError(w, http.StatusInternalServerError, "ошибка удаления поста")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "пост удалён"})
}

// getFeed обрабатывает GET /v1/feed
func (h *Handler) getFeed(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.postService.GetFeed(userID, limit, offset)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения ленты")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

// getFollowingFeed обрабатывает GET /v1/feed/following
func (h *Handler) getFollowingFeed(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.postService.GetFollowingFeed(userID, limit, offset)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения ленты подписок")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}
