package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes возвращает настроенный chi-роутер со всеми маршрутами
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(CORSMiddleware)
	r.Use(LoggingMiddleware)

	// Статические файлы (фронтенд + аватарки)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	r.Route("/v1", func(r chi.Router) {
		// Healthcheck
		r.Get("/health", h.healthCheck)

		// Авторизация (публичные)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.register)
			r.Post("/login", h.login)
			r.Post("/refresh", h.refresh)

			// Logout — защищённый
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Post("/logout", h.logout)
			})
		})

		// Лента (с опциональной авторизацией для is_liked)
		r.Group(func(r chi.Router) {
			r.Use(h.OptionalAuthMiddleware)
			r.Get("/feed", h.getFeed)
		})

		// Лента подписок — защищённая
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			r.Get("/feed/following", h.getFollowingFeed)
		})

		// Пользователи
		r.Route("/users", func(r chi.Router) {
			// /me маршруты — защищённые (должны быть ДО /{id})
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Get("/me", h.getMe)
				r.Put("/me", h.updateProfile)
				r.Post("/me/avatar", h.uploadAvatar)
			})

			// Публичные по ID
			r.Get("/{id}/followers", h.getFollowers)
			r.Get("/{id}/following", h.getFollowing)

			r.Group(func(r chi.Router) {
				r.Use(h.OptionalAuthMiddleware)
				r.Get("/{id}", h.getUser)
			})

			// Защищённые по ID
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Post("/{id}/follow", h.followUser)
				r.Delete("/{id}/follow", h.unfollowUser)
			})
		})

		// Посты
		r.Route("/posts", func(r chi.Router) {
			// Публичные (с опциональной авторизацией)
			r.Group(func(r chi.Router) {
				r.Use(h.OptionalAuthMiddleware)
				r.Get("/{id}/comments", h.getComments)
			})

			// Защищённые
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Post("/", h.createPost)
				r.Delete("/{id}", h.deletePost)
				r.Post("/{id}/comments", h.createComment)
				r.Post("/{id}/like", h.likePost)
				r.Delete("/{id}/like", h.unlikePost)
			})
		})
	})

	return r
}

// healthCheck — проверка работоспособности сервера
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
