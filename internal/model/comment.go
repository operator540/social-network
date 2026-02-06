package model

import "time"

// Comment — модель комментария
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`   // JOIN с users
	AvatarURL string    `json:"avatar_url"` // JOIN с users
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
