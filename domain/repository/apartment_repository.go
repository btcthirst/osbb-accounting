// domain/repository/apartment_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// ApartmentRepository визначає контракт для роботи з квартирами.
type ApartmentRepository interface {
	// Create створює нову квартиру
	Create(ctx context.Context, apartment *entity.Apartment) error

	// GetByID отримує квартиру за ID
	GetByID(ctx context.Context, id int64) (*entity.Apartment, error)

	// GetByNumber отримує квартиру за номером
	GetByNumber(ctx context.Context, number string) (*entity.Apartment, error)

	// Update оновлює дані квартири
	Update(ctx context.Context, apartment *entity.Apartment) error

	// SoftDelete м'яке видалення квартири
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видалену квартиру
	Restore(ctx context.Context, id int64) error

	// List отримує список квартир з фільтрацією
	List(ctx context.Context, filter ApartmentFilter) ([]*entity.Apartment, error)

	// Count отримує загальну кількість квартир
	Count(ctx context.Context, filter ApartmentFilter) (int64, error)

	// ExistsByNumber перевіряє чи існує квартира з таким номером
	ExistsByNumber(ctx context.Context, number string) (bool, error)

	// GetByFloor отримує всі квартири на поверсі
	GetByFloor(ctx context.Context, floor int) ([]*entity.Apartment, error)

	// GetByEntrance отримує всі квартири в під'їзді
	GetByEntrance(ctx context.Context, entrance int) ([]*entity.Apartment, error)

	// GetStatistics отримує статистику по квартирах
	GetStatistics(ctx context.Context) (*ApartmentStatistics, error)
}

// ApartmentFilter - критерії пошуку квартир
type ApartmentFilter struct {
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string // Пошук по номеру
	Floor          *int
	Entrance       *int
	MinArea        *float64
	MaxArea        *float64
	HasOwners      *bool // Фільтр по наявності власників
	Limit          int
	Offset         int
	OrderBy        string // "number", "floor", "area", "created_at"
	OrderDesc      bool
}

// ApartmentStatistics - статистика по квартирах
type ApartmentStatistics struct {
	TotalApartments int64
	TotalArea       float64
	AverageArea     float64
	MinArea         float64
	MaxArea         float64
	FloorCount      int
	EntranceCount   int
	WithOwners      int64
	WithoutOwners   int64
}
