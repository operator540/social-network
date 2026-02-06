package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"social-network/internal/model"
	"social-network/internal/repository"
)

var (
	ErrUserNotFound  = errors.New("пользователь не найден")
	ErrInvalidAvatar = errors.New("недопустимый формат аватарки")
)

// UserService — сервис работы с профилями
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService создаёт сервис пользователей
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile возвращает публичный профиль пользователя
func (s *UserService) GetProfile(id, currentUserID int) (*model.UserProfile, error) {
	return s.userRepo.GetProfile(id, currentUserID)
}

// GetByID возвращает пользователя по ID
func (s *UserService) GetByID(id int) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// UpdateBio обновляет bio пользователя
func (s *UserService) UpdateBio(id int, bio string) error {
	return s.userRepo.UpdateBio(id, bio)
}

// UploadAvatar сохраняет аватарку и обновляет URL в БД
func (s *UserService) UploadAvatar(userID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Проверяем MIME-тип
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", ErrInvalidAvatar
	}

	// Определяем расширение файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	// Формируем имя файла
	filename := fmt.Sprintf("avatar_%d%s", userID, ext)
	filePath := filepath.Join("web", "uploads", filename)

	// Сохраняем файл
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Обновляем URL в БД
	avatarURL := "/uploads/" + filename
	if err := s.userRepo.UpdateAvatar(userID, avatarURL); err != nil {
		return "", err
	}

	return avatarURL, nil
}
