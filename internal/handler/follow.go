package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"social-network/internal/service"
)

// followUser обрабатывает POST /v1/users/{id}/follow
func (h *Handler) followUser(w http.ResponseWriter, r *http.Request) {
	followingID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	followerID := getUserID(r)
	if err := h.followService.Follow(followerID, followingID); err != nil {
		if err == service.ErrSelfFollow {
			jsonError(w, http.StatusBadRequest, "нельзя подписаться на себя")
			return
		}
		jsonError(w, http.StatusInternalServerError, "ошибка подписки")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "вы подписались"})
}

// unfollowUser обрабатывает DELETE /v1/users/{id}/follow
func (h *Handler) unfollowUser(w http.ResponseWriter, r *http.Request) {
	followingID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	followerID := getUserID(r)
	if err := h.followService.Unfollow(followerID, followingID); err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка отписки")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "вы отписались"})
}

// getFollowers обрабатывает GET /v1/users/{id}/followers
func (h *Handler) getFollowers(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	users, err := h.followService.GetFollowers(userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения подписчиков")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// getFollowing обрабатывает GET /v1/users/{id}/following
func (h *Handler) getFollowing(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	users, err := h.followService.GetFollowing(userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения подписок")
		return
	}

	writeJSON(w, http.StatusOK, users)
}
