package service

import (
	"errors"

	"social-network/internal/model"
	"social-network/internal/repository"
)

var ErrSelfFollow = errors.New("нельзя подписаться на себя")

// FollowService — сервис работы с подписками
type FollowService struct {
	followRepo repository.FollowRepository
}

// NewFollowService создаёт сервис подписок
func NewFollowService(followRepo repository.FollowRepository) *FollowService {
	return &FollowService{followRepo: followRepo}
}

// Follow подписывает followerID на followingID
func (s *FollowService) Follow(followerID, followingID int) error {
	if followerID == followingID {
		return ErrSelfFollow
	}
	return s.followRepo.Follow(followerID, followingID)
}

// Unfollow отписывает followerID от followingID
func (s *FollowService) Unfollow(followerID, followingID int) error {
	return s.followRepo.Unfollow(followerID, followingID)
}

// GetFollowers возвращает подписчиков пользователя
func (s *FollowService) GetFollowers(userID int) ([]*model.User, error) {
	return s.followRepo.GetFollowers(userID)
}

// GetFollowing возвращает подписки пользователя
func (s *FollowService) GetFollowing(userID int) ([]*model.User, error) {
	return s.followRepo.GetFollowing(userID)
}
