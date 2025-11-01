package sqlite

import "strings"

// isUniqueConstraintError перевіряє, чи є помилка порушенням унікальності.
// Використовується для визначення конфліктів унікальних полів у SQLite.
//
// Параметри:
//   - err: помилка для перевірки
//
// Повертає:
//   - bool: true якщо це помилка UNIQUE constraint
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "UNIQUE")
}

// contains перевіряє, чи міститься підстрока в рядку.
// Допоміжна функція для роботи з текстовими повідомленнями помилок.
//
// Параметри:
//   - s: рядок для пошуку
//   - substr: підстрока, яку шукаємо
//
// Повертає:
//   - bool: true якщо substr міститься в s
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
