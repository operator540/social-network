package repository

import (
	"database/sql"

	"social-network/internal/model"
)

// commentRepo — реализация CommentRepository для PostgreSQL
type commentRepo struct {
	db *sql.DB
}

// NewCommentRepo создаёт новый репозиторий комментариев
func NewCommentRepo(db *sql.DB) CommentRepository {
	return &commentRepo{db: db}
}

func (r *commentRepo) Create(postID, userID int, content string) (*model.Comment, error) {
	comment := &model.Comment{}
	err := r.db.QueryRow(
		`INSERT INTO comments (post_id, user_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, post_id, user_id, content, created_at`,
		postID, userID, content,
	).Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Подтягиваем имя пользователя
	r.db.QueryRow(`SELECT username, avatar_url FROM users WHERE id = $1`, userID).
		Scan(&comment.Username, &comment.AvatarURL)

	return comment, nil
}

func (r *commentRepo) GetByPostID(postID int) ([]*model.Comment, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.post_id, c.user_id, u.username, u.avatar_url, c.content, c.created_at
		 FROM comments c
		 JOIN users u ON c.user_id = u.id
		 WHERE c.post_id = $1
		 ORDER BY c.created_at ASC`, postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.AvatarURL, &c.Content, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []*model.Comment{}
	}
	return comments, rows.Err()
}
