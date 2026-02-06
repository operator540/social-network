package model

import "time"

// Post — модель поста
type Post struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`    // JOIN с users
	AvatarURL  string    `json:"avatar_url"`  // JOIN с users
	Content    string    `json:"content"`
	LikesCount int       `json:"likes_count"` // Подсчёт лайков
	IsLiked    bool      `json:"is_liked"`    // Лайкнул ли текущий пользователь
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
