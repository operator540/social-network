package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

// contextKey — тип ключа для контекста
type contextKey string

const userIDKey contextKey = "user_id"

// AuthMiddleware проверяет JWT и добавляет userID в контекст
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			jsonError(w, http.StatusUnauthorized, "отсутствует токен авторизации")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			jsonError(w, http.StatusUnauthorized, "неверный формат токена")
			return
		}

		userID, err := h.authService.ParseToken(parts[1])
		if err != nil {
			jsonError(w, http.StatusUnauthorized, "невалидный токен")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware — если есть токен, добавляет userID, иначе 0
func (h *Handler) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		userID := 0

		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				if id, err := h.authService.ParseToken(parts[1]); err == nil {
					userID = id
				}
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserID извлекает userID из контекста запроса
func getUserID(r *http.Request) int {
	if id, ok := r.Context().Value(userIDKey).(int); ok {
		return id
	}
	return 0
}

// CORSMiddleware добавляет CORS-заголовки
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware логирует входящие HTTP-запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
