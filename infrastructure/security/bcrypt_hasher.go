// infrastructure/security/bcrypt_hasher.go
package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BCryptHasher реалізує service.PasswordHasher використовуючи BCrypt.
type BCryptHasher struct {
	cost int // Cost фактор для BCrypt (10-12 для production)
}

// NewBCryptHasher створює новий BCryptHasher.
// cost - складність хешування (10-12 рекомендовано, 10 за замовчуванням)
func NewBCryptHasher(cost int) *BCryptHasher {
	// Валідація cost (мінімум 4, максимум 31)
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost // 10
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost // 31
	}

	return &BCryptHasher{
		cost: cost,
	}
}

// Hash хешує пароль використовуючи BCrypt.
// Повертає хеш довжиною 60 символів.
func (h *BCryptHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Генеруємо хеш
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// BCrypt завжди повертає 60 символів
	return string(hashedBytes), nil
}

// Verify перевіряє чи відповідає пароль хешу.
// Повертає nil якщо пароль правильний, інакше помилку.
func (h *BCryptHasher) Verify(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	// BCrypt хеш завжди 60 символів
	if len(hash) != 60 {
		return fmt.Errorf("invalid hash format")
	}

	// Порівнюємо пароль з хешем
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// bcrypt.ErrMismatchedHashAndPassword означає невірний пароль
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}

// DefaultBCryptHasher повертає hasher з cost = 10 (рекомендовано для production).
func DefaultBCryptHasher() *BCryptHasher {
	return NewBCryptHasher(bcrypt.DefaultCost)
}
