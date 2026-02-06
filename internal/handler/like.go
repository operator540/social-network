package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// likePost обрабатывает POST /v1/posts/{id}/like
func (h *Handler) likePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID поста")
		return
	}

	userID := getUserID(r)
	if err := h.likeService.Like(userID, postID); err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка лайка")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "лайк поставлен"})
}

// unlikePost обрабатывает DELETE /v1/posts/{id}/like
func (h *Handler) unlikePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID поста")
		return
	}

	userID := getUserID(r)
	if err := h.likeService.Unlike(userID, postID); err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка снятия лайка")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "лайк убран"})
}
