// domain/service/password_service.go
package service

// PasswordHasher визначає контракт для хешування та перевірки паролів.
// Реалізація буде в infrastructure layer (наприклад, BCrypt).
type PasswordHasher interface {
	// Hash хешує пароль
	Hash(password string) (string, error)

	// Verify перевіряє чи відповідає пароль хешу
	Verify(password, hash string) error
}

// TokenGenerator визначає контракт для генерації токенів.
type TokenGenerator interface {
	// Generate генерує криптографічно стійкий токен
	Generate(length int) (string, error)
}
