// domain/repository/owner_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// OwnerRepository визначає контракт для роботи з власниками.
type OwnerRepository interface {
	// Create створює нового власника
	Create(ctx context.Context, owner *entity.Owner) error

	// GetByID отримує власника за ID
	GetByID(ctx context.Context, id int64) (*entity.Owner, error)

	// Update оновлює дані власника
	Update(ctx context.Context, owner *entity.Owner) error

	// SoftDelete м'яке видалення власника
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видаленого власника
	Restore(ctx context.Context, id int64) error

	// List отримує список власників з фільтрацією
	List(ctx context.Context, filter OwnerFilter) ([]*entity.Owner, error)

	// Count отримує загальну кількість власників
	Count(ctx context.Context, filter OwnerFilter) (int64, error)

	// GetByTaxNumber отримує власника за ІПН
	GetByTaxNumber(ctx context.Context, taxNumber string) (*entity.Owner, error)

	// ExistsByTaxNumber перевіряє чи існує власник з таким ІПН
	ExistsByTaxNumber(ctx context.Context, taxNumber string) (bool, error)

	// GetByApartmentID отримує всіх власників квартири (через ownership_shares)
	GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.Owner, error)

	// Search шукає власників за запитом (ПІБ, телефон, email)
	Search(ctx context.Context, query string, limit int) ([]*entity.Owner, error)
}

// OwnerFilter - критерії пошуку власників
type OwnerFilter struct {
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string // Пошук по ПІБ, телефону, email, ІПН
	HasTaxNumber   *bool  // Фільтр по наявності ІПН
	HasContact     *bool  // Фільтр по наявності контактів
	Limit          int
	Offset         int
	OrderBy        string // "name", "created_at", "updated_at"
	OrderDesc      bool
}
