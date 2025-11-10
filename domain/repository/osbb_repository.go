// domain/repository/osbb_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// OSBBRepository визначає контракт для роботи з організацією ОСББ.
// Оскільки OSBB - це singleton (один запис), методи спрощені.
type OSBBRepository interface {
	// Get отримує дані ОСББ (завжди ID = 1)
	Get(ctx context.Context) (*entity.OSBB, error)

	// Create створює організацію ОСББ (може бути викликано тільки один раз)
	Create(ctx context.Context, osbb *entity.OSBB) error

	// Update оновлює дані ОСББ
	Update(ctx context.Context, osbb *entity.OSBB) error

	// Exists перевіряє чи існує запис ОСББ
	Exists(ctx context.Context) (bool, error)

	// GetByEDRPOU отримує ОСББ за ЄДРПОУ (для валідації унікальності)
	GetByEDRPOU(ctx context.Context, edrpou string) (*entity.OSBB, error)
}
