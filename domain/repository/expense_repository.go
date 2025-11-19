// domain/repository/expense_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// ExpenseRepository визначає контракт для роботи з витратами.
type ExpenseRepository interface {
	// Create створює нову витрату
	Create(ctx context.Context, expense *entity.Expense) error

	// GetByID отримує витрату за ID
	GetByID(ctx context.Context, id int64) (*entity.Expense, error)

	// Update оновлює дані витрати
	Update(ctx context.Context, expense *entity.Expense) error

	// SoftDelete м'яке видалення витрати
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видалену витрату
	Restore(ctx context.Context, id int64) error

	// List отримує список витрат з фільтрацією
	List(ctx context.Context, filter ExpenseFilter) ([]*entity.Expense, error)

	// Count отримує загальну кількість витрат
	Count(ctx context.Context, filter ExpenseFilter) (int64, error)

	// GetByCategoryID отримує всі витрати категорії
	GetByCategoryID(ctx context.Context, categoryID int64) ([]*entity.Expense, error)

	// GetByContractorID отримує всі витрати контрагента
	GetByContractorID(ctx context.Context, contractorID int64) ([]*entity.Expense, error)

	// GetByDateRange отримує витрати за період
	GetByDateRange(ctx context.Context, startDate, endDate int64) ([]*entity.Expense, error)

	// GetStatistics отримує статистику по витратах
	GetStatistics(ctx context.Context, filter ExpenseFilter) (*ExpenseStatistics, error)

	// GetTotalByCategory отримує загальну суму витрат по категоріях
	GetTotalByCategory(ctx context.Context, startDate, endDate int64) (map[int64]float64, error)

	// GetTotalByContractor отримує загальну суму витрат по контрагентах
	GetTotalByContractor(ctx context.Context, startDate, endDate int64) (map[int64]float64, error)

	// GetTotalByMonth отримує загальну суму витрат по місяцях
	GetTotalByMonth(ctx context.Context, year int) (map[int]float64, error)
}

// ExpenseFilter - критерії пошуку витрат
type ExpenseFilter struct {
	IncludeDeleted bool
	SearchQuery    string // Пошук по опису, номеру документа
	CategoryID     *int64
	ContractorID   *int64
	PaymentStatus  *entity.PaymentStatus
	IsApproved     *bool
	ApprovedBy     *int64
	StartDate      *int64 // UNIX timestamp
	EndDate        *int64 // UNIX timestamp
	MinAmount      *float64
	MaxAmount      *float64
	DocumentType   *string
	Limit          int
	Offset         int
	OrderBy        string // "expense_date", "amount", "created_at", "updated_at"
	OrderDesc      bool
}

// ExpenseStatistics - статистика по витратах
type ExpenseStatistics struct {
	TotalExpenses    int64
	TotalAmount      float64
	PaidAmount       float64
	UnpaidAmount     float64
	PartiallyPaid    int64
	FullyPaid        int64
	Pending          int64
	Cancelled        int64
	ApprovedCount    int64
	NotApprovedCount int64
	AverageAmount    float64
	MinAmount        float64
	MaxAmount        float64
	PaymentRate      float64 // Відсоток оплачених витрат
}
