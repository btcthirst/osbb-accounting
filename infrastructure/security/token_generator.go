// infrastructure/security/token_generator.go
package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// CryptoTokenGenerator реалізує service.TokenGenerator використовуючи crypto/rand.
type CryptoTokenGenerator struct{}

// NewCryptoTokenGenerator створює новий CryptoTokenGenerator.
func NewCryptoTokenGenerator() *CryptoTokenGenerator {
	return &CryptoTokenGenerator{}
}

// Generate генерує криптографічно стійкий випадковий токен.
// length - кількість байтів (результат буде довшим через base64 encoding)
func (g *CryptoTokenGenerator) Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	// Створюємо байтовий масив
	bytes := make([]byte, length)

	// Заповнюємо випадковими байтами
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Конвертуємо в base64 URL-safe формат
	// URL-safe base64 не містить символи +, /, = що зручно для URL
	token := base64.URLEncoding.EncodeToString(bytes)

	return token, nil
}

// GenerateHex генерує токен у hex форматі (альтернатива base64).
func (g *CryptoTokenGenerator) GenerateHex(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Конвертуємо в hex (кожен байт = 2 hex символи)
	return fmt.Sprintf("%x", bytes), nil
}
