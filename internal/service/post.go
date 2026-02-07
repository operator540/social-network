package service

import (
	"errors"

	"social-network/internal/model"
	"social-network/internal/repository"
)

var (
	ErrPostNotFound = errors.New("пост не найден")
	ErrNotPostOwner = errors.New("вы не являетесь автором поста")
)

// PostService — сервис работы с постами
type PostService struct {
	postRepo repository.PostRepository
}

// NewPostService создаёт сервис постов
func NewPostService(postRepo repository.PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}

// Create создаёт новый пост
func (s *PostService) Create(userID int, content, imageURL string) (*model.Post, error) {
	return s.postRepo.Create(userID, content, imageURL)
}

// GetByID возвращает пост по ID
func (s *PostService) GetByID(id, currentUserID int) (*model.Post, error) {
	return s.postRepo.GetByID(id, currentUserID)
}

// Delete удаляет пост (только автор)
func (s *PostService) Delete(postID, userID int) error {
	post, err := s.postRepo.GetByID(postID, userID)
	if err != nil {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrNotPostOwner
	}
	return s.postRepo.Delete(postID)
}

// GetFeed возвращает глобальную ленту
func (s *PostService) GetFeed(currentUserID, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	return s.postRepo.GetFeed(currentUserID, limit, offset)
}

// GetFollowingFeed возвращает ленту подписок
func (s *PostService) GetFollowingFeed(userID, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	return s.postRepo.GetFollowingFeed(userID, limit, offset)
}

// GetByUserID возвращает посты конкретного пользователя
func (s *PostService) GetByUserID(userID, currentUserID, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	return s.postRepo.GetByUserID(userID, currentUserID, limit, offset)
}
