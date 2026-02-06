package service

import (
	"social-network/internal/model"
	"social-network/internal/repository"
)

// CommentService — сервис работы с комментариями
type CommentService struct {
	commentRepo repository.CommentRepository
}

// NewCommentService создаёт сервис комментариев
func NewCommentService(commentRepo repository.CommentRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo}
}

// Create создаёт новый комментарий
func (s *CommentService) Create(postID, userID int, content string) (*model.Comment, error) {
	return s.commentRepo.Create(postID, userID, content)
}

// GetByPostID возвращает комментарии к посту
func (s *CommentService) GetByPostID(postID int) ([]*model.Comment, error) {
	return s.commentRepo.GetByPostID(postID)
}
