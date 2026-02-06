package service

import "social-network/internal/repository"

// LikeService — сервис работы с лайками
type LikeService struct {
	likeRepo repository.LikeRepository
}

// NewLikeService создаёт сервис лайков
func NewLikeService(likeRepo repository.LikeRepository) *LikeService {
	return &LikeService{likeRepo: likeRepo}
}

// Like ставит лайк на пост
func (s *LikeService) Like(userID, postID int) error {
	return s.likeRepo.Like(userID, postID)
}

// Unlike убирает лайк с поста
func (s *LikeService) Unlike(userID, postID int) error {
	return s.likeRepo.Unlike(userID, postID)
}
