// domain/repository/contractor_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// ContractorRepository визначає контракт для роботи з контрагентами.
type ContractorRepository interface {
	// Create створює нового контрагента
	Create(ctx context.Context, contractor *entity.Contractor) error

	// GetByID отримує контрагента за ID
	GetByID(ctx context.Context, id int64) (*entity.Contractor, error)

	// GetByEDRPOU отримує контрагента за ЄДРПОУ
	GetByEDRPOU(ctx context.Context, edrpou string) (*entity.Contractor, error)

	// Update оновлює дані контрагента
	Update(ctx context.Context, contractor *entity.Contractor) error

	// SoftDelete м'яке видалення контрагента
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видаленого контрагента
	Restore(ctx context.Context, id int64) error

	// List отримує список контрагентів з фільтрацією
	List(ctx context.Context, filter ContractorFilter) ([]*entity.Contractor, error)

	// Count отримує загальну кількість контрагентів
	Count(ctx context.Context, filter ContractorFilter) (int64, error)

	// ExistsByEDRPOU перевіряє чи існує контрагент з таким ЄДРПОУ
	ExistsByEDRPOU(ctx context.Context, edrpou string) (bool, error)

	// Search шукає контрагентів за запитом (назва, ЄДРПОУ, контактна особа)
	Search(ctx context.Context, query string, limit int) ([]*entity.Contractor, error)

	// GetByType отримує всіх контрагентів певного типу
	GetByType(ctx context.Context, contractorType entity.ContractorType) ([]*entity.Contractor, error)
}

// ContractorFilter - критерії пошуку контрагентів
type ContractorFilter struct {
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string                 // Пошук по назві, ЄДРПОУ, контактній особі
	ContractorType *entity.ContractorType // Фільтр по типу
	HasBankDetails *bool                  // Фільтр по наявності банківських реквізитів
	HasContract    *bool                  // Фільтр по наявності договору
	Limit          int
	Offset         int
	OrderBy        string // "name", "type", "created_at", "updated_at"
	OrderDesc      bool
}
