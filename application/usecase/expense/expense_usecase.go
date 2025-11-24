// application/usecase/expense/expense_usecase.go
package expense

import (
	"time"

	"osbb-accounting/domain/entity"
)

// ExpenseOutput - DTO для відображення витрати
type ExpenseOutput struct {
	ID             int64
	CategoryID     int64
	CategoryName   string
	ContractorID   *int64
	ContractorName *string
	ExpenseDate    time.Time
	Amount         float64
	Description    string

	// Документ
	DocumentType   *string
	DocumentNumber *string
	DocumentDate   *time.Time

	// Статус
	PaymentStatus     string
	PaymentStatusName string
	PaidAmount        float64
	PaymentDate       *time.Time

	Notes *string

	IsApproved bool
	ApprovedBy *int64
	ApprovedAt *time.Time
}

// CreateExpenseInput - вхідні дані для створення витрати
type CreateExpenseInput struct {
	CurrentUserID int64
	CategoryID    int64
	ContractorID  *int64
	ExpenseDate   time.Time
	Amount        float64
	Description   string

	// Опціонально
	DocumentType   *entity.DocumentType
	DocumentNumber *string
	DocumentDate   *time.Time
	Notes          *string
}

// UpdateExpenseInput - вхідні дані для оновлення витрати
type UpdateExpenseInput struct {
	CurrentUserID int64
	ExpenseID     int64
	CategoryID    int64
	ContractorID  *int64
	ExpenseDate   time.Time
	Amount        float64
	Description   string
	Notes         *string
}

// DeleteExpenseInput - вхідні дані для видалення витрати
type DeleteExpenseInput struct {
	CurrentUserID int64
	ExpenseID     int64
}

// ApproveExpenseInput - вхідні дані для затвердження
type ApproveExpenseInput struct {
	CurrentUserID int64
	ExpenseID     int64
}

// UnapproveExpenseInput - вхідні дані для скасування затвердження
type UnapproveExpenseInput struct {
	CurrentUserID int64
	ExpenseID     int64
}

// ListExpensesInput - вхідні дані для списку
type ListExpensesInput struct {
	CurrentUserID int64
	Limit         int
	Offset        int
	CategoryID    *int64
	ContractorID  *int64
	PaymentStatus *entity.PaymentStatus
	IsApproved    *bool
	StartDate     *int64
	EndDate       *int64
	SearchQuery   string
}

// ListExpensesOutput - результат списку
type ListExpensesOutput struct {
	Expenses []*ExpenseOutput
	Total    int64
}

// GetExpenseStatisticsInput - вхідні дані для статистики
type GetExpenseStatisticsInput struct {
	CurrentUserID int64
	StartDate     *int64
	EndDate       *int64
}
