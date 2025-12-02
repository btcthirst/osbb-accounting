// application/usecase/expense/expense_output.go
package expense

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
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

// toOutput converts entity to output DTO
func toOutput(e *entity.Expense, categoryRepo repository.ExpenseCategoryRepository) *ExpenseOutput {
	// Get category name
	categoryName := fmt.Sprintf("Category %d", e.CategoryID) // Fallback
	if category, err := categoryRepo.GetByID(context.Background(), e.CategoryID); err == nil {
		categoryName = category.Name
	}

	return &ExpenseOutput{
		ID:                e.ID,
		CategoryID:        e.CategoryID,
		CategoryName:      categoryName,
		ContractorID:      e.ContractorID,
		ExpenseDate:       e.ExpenseDate,
		Amount:            e.Amount,
		Description:       e.Description,
		DocumentType:      e.DocumentType,
		DocumentNumber:    e.DocumentNumber,
		DocumentDate:      e.DocumentDate,
		PaymentStatus:     string(e.PaymentStatus),
		PaymentStatusName: e.PaymentStatus.GetDisplayName(),
		PaidAmount:        e.PaidAmount,
		PaymentDate:       e.PaymentDate,
		Notes:             e.Notes,
		IsApproved:        e.IsApproved(),
		ApprovedBy:        e.ApprovedBy,
		ApprovedAt:        e.ApprovedAt,
	}
}
