package handler

import (
	"log"
	"net/http"

	"social-network/internal/service"
)

// registerRequest — тело запроса регистрации
type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest — тело запроса логина
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// refreshRequest — тело запроса обновления токена
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// register обрабатывает POST /v1/auth/register
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "все поля обязательны")
		return
	}

	if len(req.Password) < 6 {
		jsonError(w, http.StatusBadRequest, "пароль должен быть не менее 6 символов")
		return
	}

	user, tokens, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserExists {
			jsonError(w, http.StatusConflict, "пользователь уже существует")
			return
		}
		jsonError(w, http.StatusInternalServerError, "ошибка регистрации")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// login обрабатывает POST /v1/auth/login
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "email и пароль обязательны")
		return
	}

	user, tokens, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			jsonError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		log.Printf("Ошибка логина: %v", err)
		jsonError(w, http.StatusInternalServerError, "ошибка входа")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// refresh обрабатывает POST /v1/auth/refresh
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.RefreshToken == "" {
		jsonError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "невалидный refresh-токен")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// logout обрабатывает POST /v1/auth/logout
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := readJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.RefreshToken == "" {
		jsonError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка выхода")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "вы вышли из системы"})
}
