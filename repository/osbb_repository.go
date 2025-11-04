package repository

import "osbb-accounting/domain"

// OSBBRepository визначає контракт для роботи з організацією ОСББ.
// Оскільки ОСББ завжди одна (ID = 1), тут немає методів Create/Delete.
type OSBBRepository interface {
	// Get повертає дані ОСББ (завжди ID = 1).
	Get() (*domain.OSBB, error)

	// Update оновлює дані ОСББ.
	Update(osbb *domain.OSBB) error

	// UpdateRates оновлює тарифи.
	UpdateRates(maintenanceRate, utilitiesRate float64) error
}
