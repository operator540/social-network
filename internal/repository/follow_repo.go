package repository

import (
	"database/sql"

	"social-network/internal/model"
)

// followRepo — реализация FollowRepository для PostgreSQL
type followRepo struct {
	db *sql.DB
}

// NewFollowRepo создаёт новый репозиторий подписок
func NewFollowRepo(db *sql.DB) FollowRepository {
	return &followRepo{db: db}
}

func (r *followRepo) Follow(followerID, followingID int) error {
	_, err := r.db.Exec(
		`INSERT INTO follows (follower_id, following_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		followerID, followingID,
	)
	return err
}

func (r *followRepo) Unfollow(followerID, followingID int) error {
	_, err := r.db.Exec(
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	return err
}

func (r *followRepo) GetFollowers(userID int) ([]*model.User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.username, u.email, u.bio, u.avatar_url, u.created_at, u.updated_at
		 FROM users u
		 JOIN follows f ON u.id = f.follower_id
		 WHERE f.following_id = $1
		 ORDER BY f.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

func (r *followRepo) GetFollowing(userID int) ([]*model.User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.username, u.email, u.bio, u.avatar_url, u.created_at, u.updated_at
		 FROM users u
		 JOIN follows f ON u.id = f.following_id
		 WHERE f.follower_id = $1
		 ORDER BY f.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

func (r *followRepo) IsFollowing(followerID, followingID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&exists)
	return exists, err
}

// scanUsers сканирует строки результата в слайс пользователей
func scanUsers(rows *sql.Rows) ([]*model.User, error) {
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Bio, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []*model.User{}
	}
	return users, rows.Err()
}
