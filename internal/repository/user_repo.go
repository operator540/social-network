package repository

import (
	"database/sql"

	"social-network/internal/model"
)

// userRepo — реализация UserRepository для PostgreSQL
type userRepo struct {
	db *sql.DB
}

// NewUserRepo создаёт новый репозиторий пользователей
func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(username, email, passwordHash string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, email, bio, avatar_url, created_at, updated_at`,
		username, email, passwordHash,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) GetByID(id int) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, bio, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) GetByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, bio, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) GetByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, bio, avatar_url, created_at, updated_at
		 FROM users WHERE username = $1`, username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Bio, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) UpdateBio(id int, bio string) error {
	_, err := r.db.Exec(
		`UPDATE users SET bio = $1, updated_at = NOW() WHERE id = $2`, bio, id,
	)
	return err
}

func (r *userRepo) UpdateAvatar(id int, avatarURL string) error {
	_, err := r.db.Exec(
		`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`, avatarURL, id,
	)
	return err
}

func (r *userRepo) GetProfile(id, currentUserID int) (*model.UserProfile, error) {
	profile := &model.UserProfile{}
	err := r.db.QueryRow(
		`SELECT u.id, u.username, u.email, u.bio, u.avatar_url, u.created_at, u.updated_at,
			(SELECT COUNT(*) FROM follows WHERE following_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = u.id) as is_following
		 FROM users u WHERE u.id = $1`, id, currentUserID,
	).Scan(
		&profile.ID, &profile.Username, &profile.Email, &profile.Bio,
		&profile.AvatarURL, &profile.CreatedAt, &profile.UpdatedAt,
		&profile.FollowersCount, &profile.FollowingCount, &profile.IsFollowing,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
