package repository

import (
	"database/sql"

	"social-network/internal/model"
)

// tokenRepo — реализация TokenRepository для PostgreSQL
type tokenRepo struct {
	db *sql.DB
}

// NewTokenRepo создаёт новый репозиторий токенов
func NewTokenRepo(db *sql.DB) TokenRepository {
	return &tokenRepo{db: db}
}

func (r *tokenRepo) Create(userID int, tokenHash string, expiresAt any) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *tokenRepo) GetByHash(tokenHash string) (*model.RefreshToken, error) {
	token := &model.RefreshToken{}
	err := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *tokenRepo) DeleteByHash(tokenHash string) error {
	_, err := r.db.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *tokenRepo) DeleteByUserID(userID int) error {
	_, err := r.db.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}
