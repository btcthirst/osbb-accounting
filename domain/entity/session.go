// domain/entity/session.go
package entity

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

// Session представляє активну сесію користувача.
// Використовується для stateful аутентифікації (альтернатива JWT).
type Session struct {
	ID        int64
	UserID    int64
	Token     string // Унікальний токен сесії
	ExpiresAt time.Time
	CreatedAt time.Time
	IPAddress *string
	UserAgent *string
}

// Session configuration constants
const (
	SessionTokenLength = 32                  // Довжина токену в байтах
	SessionDuration    = 24 * time.Hour      // Час життя сесії за замовчуванням
	SessionMaxDuration = 30 * 24 * time.Hour // Максимальний час життя сесії
)

var (
	ErrSessionUserIDRequired = errors.New("user ID is required")
	ErrSessionExpired        = errors.New("session has expired")
	ErrSessionTokenInvalid   = errors.New("invalid session token")
	ErrSessionGenerateToken  = errors.New("failed to generate session token")
)

// NewSession створює нову сесію для користувача.
func NewSession(userID int64, duration time.Duration, ipAddress, userAgent *string) (*Session, error) {
	if userID <= 0 {
		return nil, ErrSessionUserIDRequired
	}

	// Генеруємо криптографічно стійкий токен
	token, err := generateSecureToken(SessionTokenLength)
	if err != nil {
		return nil, ErrSessionGenerateToken
	}

	// Обмежуємо максимальний час життя
	if duration > SessionMaxDuration {
		duration = SessionMaxDuration
	}
	if duration <= 0 {
		duration = SessionDuration
	}

	now := time.Now()

	return &Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(duration),
		CreatedAt: now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// IsValid перевіряє чи є сесія дійсною (не закінчилася).
func (s *Session) IsValid() error {
	if time.Now().After(s.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}

// IsExpired перевіряє чи закінчилася сесія.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Extend продовжує час життя сесії.
func (s *Session) Extend(duration time.Duration) {
	if duration > SessionMaxDuration {
		duration = SessionMaxDuration
	}
	s.ExpiresAt = time.Now().Add(duration)
}

// RemainingTime повертає час до закінчення сесії.
func (s *Session) RemainingTime() time.Duration {
	remaining := time.Until(s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// generateSecureToken генерує криптографічно стійкий випадковий токен.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Конвертуємо в base64 URL-safe формат
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SessionCleanupPolicy визначає політику очищення старих сесій.
type SessionCleanupPolicy struct {
	ExpiredSessionsRetentionDays int           // Скільки днів зберігати закінчені сесії
	CleanupInterval              time.Duration // Як часто запускати очищення
}

// DefaultSessionCleanupPolicy повертає політику очищення за замовчуванням.
func DefaultSessionCleanupPolicy() SessionCleanupPolicy {
	return SessionCleanupPolicy{
		ExpiredSessionsRetentionDays: 7,              // Зберігати 7 днів для аудиту
		CleanupInterval:              24 * time.Hour, // Очищати раз на добу
	}
}
