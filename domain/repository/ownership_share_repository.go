// domain/repository/ownership_share_repository.go
package repository

import (
	"context"

	"osbb-accounting/domain/entity"
)

// OwnershipShareRepository визначає контракт для роботи з частками власності.
type OwnershipShareRepository interface {
	// Create створює нову частку власності
	Create(ctx context.Context, share *entity.OwnershipShare) error

	// GetByID отримує частку за ID
	GetByID(ctx context.Context, id int64) (*entity.OwnershipShare, error)

	// Update оновлює частку власності
	Update(ctx context.Context, share *entity.OwnershipShare) error

	// SoftDelete м'яке видалення частки
	SoftDelete(ctx context.Context, id int64) error

	// List отримує список часток з фільтрацією
	List(ctx context.Context, filter OwnershipShareFilter) ([]*entity.OwnershipShare, error)

	// Count отримує загальну кількість часток
	Count(ctx context.Context, filter OwnershipShareFilter) (int64, error)

	// GetByOwnerID отримує всі частки власника
	GetByOwnerID(ctx context.Context, ownerID int64) ([]*entity.OwnershipShare, error)

	// GetByApartmentID отримує всі частки квартири
	GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.OwnershipShare, error)

	// GetActiveByApartment отримує тільки активні частки квартири
	GetActiveByApartment(ctx context.Context, apartmentID int64) ([]*entity.OwnershipShare, error)

	// CheckDuplicateActiveOwnership перевіряє наявність активної частки
	CheckDuplicateActiveOwnership(ctx context.Context, ownerID, apartmentID int64, excludeID *int64) (bool, error)

	// CalculateTotalShareForApartment обчислює загальну суму часток у квартирі
	CalculateTotalShareForApartment(ctx context.Context, apartmentID int64, excludeID *int64) (float64, error)

	// GetWithDetails отримує частку з повною інформацією (owner + apartment)
	GetWithDetails(ctx context.Context, id int64) (*OwnershipShareDetails, error)

	// ListWithDetails отримує список часток з повною інформацією
	ListWithDetails(ctx context.Context, filter OwnershipShareFilter) ([]*OwnershipShareDetails, error)
}

// OwnershipShareFilter - критерії пошуку часток власності
type OwnershipShareFilter struct {
	IncludeDeleted    bool
	IsActive          *bool
	OwnerID           *int64
	ApartmentID       *int64
	OwnershipType     *entity.OwnershipType
	StartDateFrom     *int64 // UNIX timestamp
	StartDateTo       *int64
	IsCurrentlyActive bool // Фільтр по поточній активності (враховує дати)
	Limit             int
	Offset            int
	OrderBy           string // "start_date", "owner", "apartment"
	OrderDesc         bool
}

// OwnershipShareDetails - частка з повною інформацією
type OwnershipShareDetails struct {
	Share             *entity.OwnershipShare
	OwnerName         string
	OwnerPhone        *string
	OwnerEmail        *string
	ApartmentNumber   string
	ApartmentFloor    int
	ApartmentEntrance *int
}
