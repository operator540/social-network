package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"social-network/internal/model"
	"social-network/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	ErrUserExists         = errors.New("пользователь уже существует")
	ErrInvalidToken       = errors.New("невалидный токен")
	ErrTokenExpired       = errors.New("токен истёк")
)

// AuthService — сервис авторизации
type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwtSecret []byte
}

// NewAuthService создаёт сервис авторизации
func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(username, email, password string) (*model.User, *model.TokenPair, error) {
	// Проверяем, не занят ли email
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, nil, ErrUserExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	// Проверяем, не занят ли username
	_, err = s.userRepo.GetByUsername(username)
	if err == nil {
		return nil, nil, ErrUserExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	// Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// Создаём пользователя
	user, err := s.userRepo.Create(username, email, string(hash))
	if err != nil {
		return nil, nil, err
	}

	// Генерируем токены
	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login аутентифицирует пользователя по email и паролю
func (s *AuthService) Login(email, password string) (*model.User, *model.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Сравниваем пароль с хешем
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Refresh обновляет access-токен по refresh-токену
func (s *AuthService) Refresh(refreshToken string) (*model.TokenPair, error) {
	// Хешируем токен для поиска в БД
	hash := hashToken(refreshToken)

	stored, err := s.tokenRepo.GetByHash(hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	// Проверяем срок действия
	if time.Now().After(stored.ExpiresAt) {
		s.tokenRepo.DeleteByHash(hash)
		return nil, ErrTokenExpired
	}

	// Удаляем старый refresh-токен
	s.tokenRepo.DeleteByHash(hash)

	// Генерируем новую пару токенов
	return s.generateTokens(stored.UserID)
}

// Logout инвалидирует refresh-токен
func (s *AuthService) Logout(refreshToken string) error {
	hash := hashToken(refreshToken)
	return s.tokenRepo.DeleteByHash(hash)
}

// ParseToken парсит и валидирует access JWT-токен, возвращает userID
func (s *AuthService) ParseToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}

	return int(userIDFloat), nil
}

// generateTokens генерирует пару access + refresh токенов
func (s *AuthService) generateTokens(userID int) (*model.TokenPair, error) {
	// Access-токен — 15 минут
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh-токен — 7 дней (с уникальным jti для уникальности)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
		"jti":     generateJTI(),
	}
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshJWT.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Сохраняем хеш refresh-токена в БД
	hash := hashToken(refreshString)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.Create(userID, hash, expiresAt); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
	}, nil
}

// generateJTI генерирует уникальный идентификатор токена
func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// hashToken хеширует токен через SHA256
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
