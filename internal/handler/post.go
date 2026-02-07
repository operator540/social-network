package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"social-network/internal/service"
)

// createPost обрабатывает POST /v1/posts
// Принимает multipart/form-data с полями: content (текст), image (файл, необязательно)
func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	contentType := r.Header.Get("Content-Type")

	var content string
	var imageURL string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Multipart — может содержать изображение
		r.ParseMultipartForm(10 << 20) // 10 MB

		content = r.FormValue("content")

		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Проверяем MIME-тип
			ct := header.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "image/") {
				jsonError(w, http.StatusBadRequest, "допускаются только изображения")
				return
			}

			// Генерируем уникальное имя файла
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".jpg"
			}
			filename := fmt.Sprintf("post_%d_%d%s", userID, time.Now().UnixNano(), ext)
			filePath := filepath.Join("web", "uploads", filename)

			dst, err := os.Create(filePath)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "ошибка сохранения изображения")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				jsonError(w, http.StatusInternalServerError, "ошибка сохранения изображения")
				return
			}

			imageURL = "/uploads/" + filename
		}
	} else {
		// JSON-запрос (обратная совместимость)
		var req struct {
			Content string `json:"content"`
		}
		if err := readJSON(r, &req); err != nil {
			jsonError(w, http.StatusBadRequest, "неверный формат запроса")
			return
		}
		content = req.Content
	}

	if content == "" && imageURL == "" {
		jsonError(w, http.StatusBadRequest, "текст или изображение обязательны")
		return
	}

	post, err := h.postService.Create(userID, content, imageURL)
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

// getUserPosts обрабатывает GET /v1/users/{id}/posts
func (h *Handler) getUserPosts(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "неверный ID пользователя")
		return
	}

	currentUserID := getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.postService.GetByUserID(userID, currentUserID, limit, offset)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "ошибка получения постов пользователя")
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
