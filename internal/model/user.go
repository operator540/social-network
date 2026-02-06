package model

import "time"

// User — модель пользователя
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Никогда не отдаём в JSON
	Bio          string    `json:"bio"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserProfile — публичный профиль с подсчётом подписчиков
type UserProfile struct {
	User
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
	IsFollowing    bool `json:"is_following"` // Подписан ли текущий пользователь
}
