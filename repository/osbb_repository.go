package repository

import "osbb-accounting/domain"

// =============================================================================
// OSBBRepository - управління даними ОСББ
// =============================================================================

// OSBBRepository визначає контракт для операцій з ОСББ.
// В системі може бути тільки одна організація ОСББ.
type OSBBRepository interface {
	// Save зберігає дані ОСББ в БД.
	// Якщо ОСББ вже існує, повертає domain.ErrOSBBAlreadyExists.
	//
	// Параметри:
	//   - osbb: дані ОСББ для збереження
	//
	// Повертає:
	//   - error: nil при успіху
	Save(osbb *domain.OSBB) error

	// Get отримує дані ОСББ.
	// В системі має бути тільки один запис ОСББ.
	//
	// Повертає:
	//   - *domain.OSBB: дані ОСББ
	//   - error: domain.ErrOSBBNotFound якщо не знайдено
	Get() (*domain.OSBB, error)

	// Update оновлює дані ОСББ.
	//
	// Параметри:
	//   - osbb: оновлені дані
	//
	// Повертає:
	//   - error: nil при успіху
	Update(osbb *domain.OSBB) error

	// Exists перевіряє, чи існує ОСББ в системі.
	//
	// Повертає:
	//   - bool: true якщо ОСББ існує
	//   - error: nil при успіху
	Exists() (bool, error)

	// GetSettings отримує налаштування та статистику ОСББ.
	//
	// Повертає:
	//   - *domain.OSBBSettings: налаштування з статистикою
	//   - error: nil при успіху
	GetSettings() (*domain.OSBBSettings, error)
}
