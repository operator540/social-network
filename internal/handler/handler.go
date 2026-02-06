package handler

import "social-network/internal/service"

// Handler — главная структура, объединяющая все сервисы
type Handler struct {
	authService    *service.AuthService
	userService    *service.UserService
	postService    *service.PostService
	commentService *service.CommentService
	followService  *service.FollowService
	likeService    *service.LikeService
}

// NewHandler создаёт новый Handler с внедрёнными зависимостями
func NewHandler(
	authService *service.AuthService,
	userService *service.UserService,
	postService *service.PostService,
	commentService *service.CommentService,
	followService *service.FollowService,
	likeService *service.LikeService,
) *Handler {
	return &Handler{
		authService:    authService,
		userService:    userService,
		postService:    postService,
		commentService: commentService,
		followService:  followService,
		likeService:    likeService,
	}
}
