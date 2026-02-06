package repository

import "social-network/internal/model"

// UserRepository — интерфейс работы с пользователями
type UserRepository interface {
	Create(username, email, passwordHash string) (*model.User, error)
	GetByID(id int) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	UpdateBio(id int, bio string) error
	UpdateAvatar(id int, avatarURL string) error
	GetProfile(id, currentUserID int) (*model.UserProfile, error)
}

// PostRepository — интерфейс работы с постами
type PostRepository interface {
	Create(userID int, content string) (*model.Post, error)
	GetByID(id, currentUserID int) (*model.Post, error)
	Delete(id int) error
	GetFeed(currentUserID, limit, offset int) ([]*model.Post, error)
	GetFollowingFeed(userID, limit, offset int) ([]*model.Post, error)
}

// CommentRepository — интерфейс работы с комментариями
type CommentRepository interface {
	Create(postID, userID int, content string) (*model.Comment, error)
	GetByPostID(postID int) ([]*model.Comment, error)
}

// FollowRepository — интерфейс работы с подписками
type FollowRepository interface {
	Follow(followerID, followingID int) error
	Unfollow(followerID, followingID int) error
	GetFollowers(userID int) ([]*model.User, error)
	GetFollowing(userID int) ([]*model.User, error)
	IsFollowing(followerID, followingID int) (bool, error)
}

// LikeRepository — интерфейс работы с лайками
type LikeRepository interface {
	Like(userID, postID int) error
	Unlike(userID, postID int) error
	IsLiked(userID, postID int) (bool, error)
}

// TokenRepository — интерфейс работы с refresh-токенами
type TokenRepository interface {
	Create(userID int, tokenHash string, expiresAt any) error
	GetByHash(tokenHash string) (*model.RefreshToken, error)
	DeleteByHash(tokenHash string) error
	DeleteByUserID(userID int) error
}
