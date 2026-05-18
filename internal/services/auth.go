package services

import (
	"errors"
	"strings"
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
	return s.LoginWithContext(username, password, "", "")
}

func (s *AuthService) LoginWithContext(username, password, ip, userAgent string) (*models.User, error) {
	username = strings.TrimSpace(username)
	var attempts int64
	s.db.Model(&models.LoginAttempt{}).
		Where("(username = ? OR ip = ?) AND success = ? AND created_at > ?", username, ip, false, time.Now().Add(-15*time.Minute)).
		Count(&attempts)
	if attempts >= 5 {
		s.auditSecurity(nil, "brute_force_blocked", "warning", ip, userAgent, username)
		return nil, errors.New("too many login attempts, try again later")
	}
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		s.recordAttempt(username, ip, false)
		s.auditSecurity(nil, "login_failed", "warning", ip, userAgent, username)
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.recordAttempt(username, ip, false)
		s.auditSecurity(user.TenantID, "login_failed", "warning", ip, userAgent, username)
		return nil, errors.New("invalid credentials")
	}
	s.recordAttempt(username, ip, true)
	s.auditSecurity(user.TenantID, "login_success", "info", ip, userAgent, username)
	return &user, nil
}

func ValidatePasswordPolicy(password string) error {
	if len(password) < 10 {
		return errors.New("password must be at least 10 characters")
	}
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return errors.New("password must include uppercase, lowercase, number, and symbol")
	}
	return nil
}

func (s *AuthService) recordAttempt(username, ip string, success bool) {
	_ = s.db.Create(&models.LoginAttempt{Username: username, IP: ip, Success: success}).Error
}

func (s *AuthService) auditSecurity(tenantID *uint, eventType, severity, ip, userAgent, details string) {
	_ = s.db.Create(&models.SecurityEvent{TenantID: tenantID, Type: eventType, Severity: severity, IP: ip, UserAgent: userAgent, Details: details}).Error
}
