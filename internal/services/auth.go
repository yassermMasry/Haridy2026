package services

import (
	"errors"
	"time"

	"haridy2026/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(username, password string) (*models.User, error) {
	var attempts int64
	s.db.Model(&models.LoginAttempt{}).Where("username = ? AND success = ? AND created_at > ?", username, false, time.Now().Add(-15*time.Minute)).Count(&attempts)
	if attempts >= 5 {
		return nil, errors.New("too many login attempts, try again later")
	}
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		s.db.Create(&models.LoginAttempt{Username: username, Success: false})
		return nil, errors.New("بيانات الدخول غير صحيحة")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.db.Create(&models.LoginAttempt{Username: username, Success: false})
		return nil, errors.New("بيانات الدخول غير صحيحة")
	}
	s.db.Create(&models.LoginAttempt{Username: username, Success: true})
	return &user, nil
}
