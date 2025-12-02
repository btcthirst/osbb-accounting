package expense

import (
	"context"
	"fmt"

	"osbb-accounting/domain/repository"
)

// GetExpenseUseCase gets an expense by ID.
type GetExpenseUseCase struct {
	expenseRepo  repository.ExpenseRepository
	categoryRepo repository.ExpenseCategoryRepository
}

// NewGetExpenseUseCase creates a new use case.
func NewGetExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *GetExpenseUseCase {
	return &GetExpenseUseCase{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// Execute executes the retrieval of an expense.
func (uc *GetExpenseUseCase) Execute(ctx context.Context, id int64) (*ExpenseOutput, error) {
	e, err := uc.expenseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return toOutput(e, uc.categoryRepo), nil
}
