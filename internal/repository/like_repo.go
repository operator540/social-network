package repository

import "database/sql"

// likeRepo — реализация LikeRepository для PostgreSQL
type likeRepo struct {
	db *sql.DB
}

// NewLikeRepo создаёт новый репозиторий лайков
func NewLikeRepo(db *sql.DB) LikeRepository {
	return &likeRepo{db: db}
}

func (r *likeRepo) Like(userID, postID int) error {
	_, err := r.db.Exec(
		`INSERT INTO likes (user_id, post_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	return err
}

func (r *likeRepo) Unlike(userID, postID int) error {
	_, err := r.db.Exec(
		`DELETE FROM likes WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	return err
}

func (r *likeRepo) IsLiked(userID, postID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	return exists, err
}
