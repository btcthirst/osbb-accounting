// domain/repository/expense_category_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// ExpenseCategoryRepository визначає контракт для роботи з категоріями витрат.
type ExpenseCategoryRepository interface {
	// Create створює нову категорію
	Create(ctx context.Context, category *entity.ExpenseCategory) error

	// GetByID отримує категорію за ID
	GetByID(ctx context.Context, id int64) (*entity.ExpenseCategory, error)

	// GetByCode отримує категорію за кодом
	GetByCode(ctx context.Context, code string) (*entity.ExpenseCategory, error)

	// Update оновлює дані категорії
	Update(ctx context.Context, category *entity.ExpenseCategory) error

	// SoftDelete м'яке видалення категорії
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видалену категорію
	Restore(ctx context.Context, id int64) error

	// List отримує список категорій з фільтрацією
	List(ctx context.Context, filter ExpenseCategoryFilter) ([]*entity.ExpenseCategory, error)

	// Count отримує загальну кількість категорій
	Count(ctx context.Context, filter ExpenseCategoryFilter) (int64, error)

	// GetRootCategories отримує всі кореневі категорії (без батьківської)
	GetRootCategories(ctx context.Context) ([]*entity.ExpenseCategory, error)

	// GetChildren отримує всі дочірні категорії
	GetChildren(ctx context.Context, parentID int64) ([]*entity.ExpenseCategory, error)

	// GetTree отримує повне дерево категорій
	GetTree(ctx context.Context) ([]*entity.ExpenseCategory, error)

	// GetAncestors отримує всіх предків категорії (шлях до кореня)
	GetAncestors(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error)

	// GetDescendants отримує всіх нащадків категорії (все піддерево)
	GetDescendants(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error)

	// HasChildren перевіряє чи має категорія дочірні елементи
	HasChildren(ctx context.Context, categoryID int64) (bool, error)

	// ExistsByCode перевіряє чи існує категорія з таким кодом
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// GetDepth отримує глибину категорії в дереві
	GetDepth(ctx context.Context, categoryID int64) (int, error)

	// Move переміщує категорію до іншого батька
	Move(ctx context.Context, categoryID int64, newParentID *int64) error

	// GetByType отримує всі категорії певного типу
	GetByType(ctx context.Context, categoryType entity.CategoryType) ([]*entity.ExpenseCategory, error)
}

// ExpenseCategoryFilter - критерії пошуку категорій
type ExpenseCategoryFilter struct {
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string               // Пошук по назві або коду
	CategoryType   *entity.CategoryType // Фільтр по типу
	ParentID       *int64               // Фільтр по батьківській категорії (NULL для кореневих)
	OnlyRoot       bool                 // Тільки кореневі категорії
	MaxDepth       *int                 // Максимальна глибина
	Limit          int
	Offset         int
	OrderBy        string // "name", "code", "type", "created_at"
	OrderDesc      bool
}
