package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// updateProfileRequest — тело запроса обновления профиля
type updateProfileRequest struct {
	Bio string `json:"bio"`
}

// getMe обрабатывает GET /v1/users/me
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	user, err := h.userService.GetByID(userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения профиля")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// getUser обрабатывает GET /v1/users/{id}
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	currentUserID := getUserID(r)
	profile, err := h.userService.GetProfile(id, currentUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "пользователь не найден")
			return
		}
		jsonError(w, http.StatusInternalServerError, "ошибка получения профиля")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// updateProfile обрабатывает PUT /v1/users/me
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	userID := getUserID(r)
	if err := h.userService.UpdateBio(userID, req.Bio); err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка обновления профиля")
		return
	}

	// Возвращаем обновлённого пользователя
	user, _ := h.userService.GetByID(userID)
	writeJSON(w, http.StatusOK, user)
}

// uploadAvatar обрабатывает POST /v1/users/me/avatar
func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер файла до 5 МБ
	r.ParseMultipartForm(5 << 20)

	file, header, err := r.FormFile("avatar")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "файл аватарки не найден")
		return
	}
	defer file.Close()

	userID := getUserID(r)
	avatarURL, err := h.userService.UploadAvatar(userID, file, header)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ошибка загрузки аватарки")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"avatar_url": avatarURL})
}
