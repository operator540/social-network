package repository

import (
	"database/sql"

	"social-network/internal/model"
)

// postRepo — реализация PostRepository для PostgreSQL
type postRepo struct {
	db *sql.DB
}

// NewPostRepo создаёт новый репозиторий постов
func NewPostRepo(db *sql.DB) PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(userID int, content, imageURL string) (*model.Post, error) {
	post := &model.Post{}
	err := r.db.QueryRow(
		`INSERT INTO posts (user_id, content, image_url)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, content, image_url, created_at, updated_at`,
		userID, content, imageURL,
	).Scan(&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Подтягиваем имя пользователя
	r.db.QueryRow(`SELECT username, avatar_url FROM users WHERE id = $1`, userID).
		Scan(&post.Username, &post.AvatarURL)

	return post, nil
}

func (r *postRepo) GetByID(id, currentUserID int) (*model.Post, error) {
	post := &model.Post{}
	err := r.db.QueryRow(
		`SELECT p.id, p.user_id, u.username, u.avatar_url, p.content, p.image_url,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) as likes_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $2) as is_liked,
			p.created_at, p.updated_at
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.id = $1`, id, currentUserID,
	).Scan(&post.ID, &post.UserID, &post.Username, &post.AvatarURL, &post.Content, &post.ImageURL,
		&post.LikesCount, &post.IsLiked, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *postRepo) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM posts WHERE id = $1`, id)
	return err
}

func (r *postRepo) GetFeed(currentUserID, limit, offset int) ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.user_id, u.username, u.avatar_url, p.content, p.image_url,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) as likes_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) as is_liked,
			p.created_at, p.updated_at
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`, currentUserID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (r *postRepo) GetFollowingFeed(userID, limit, offset int) ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.user_id, u.username, u.avatar_url, p.content, p.image_url,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) as likes_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) as is_liked,
			p.created_at, p.updated_at
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.user_id IN (SELECT following_id FROM follows WHERE follower_id = $1)
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`, userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (r *postRepo) GetByUserID(userID, currentUserID, limit, offset int) ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.user_id, u.username, u.avatar_url, p.content, p.image_url,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) as likes_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $2) as is_liked,
			p.created_at, p.updated_at
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.user_id = $1
		 ORDER BY p.created_at DESC
		 LIMIT $3 OFFSET $4`, userID, currentUserID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}

// scanPosts сканирует строки результата в слайс постов
func scanPosts(rows *sql.Rows) ([]*model.Post, error) {
	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(&post.ID, &post.UserID, &post.Username, &post.AvatarURL,
			&post.Content, &post.ImageURL, &post.LikesCount, &post.IsLiked, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if posts == nil {
		posts = []*model.Post{}
	}
	return posts, rows.Err()
}
