package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListExpensesUseCase lists expenses.
type ListExpensesUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
	categoryRepo   repository.ExpenseCategoryRepository
}

// NewListExpensesUseCase creates a new use case.
func NewListExpensesUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *ListExpensesUseCase {
	return &ListExpensesUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
		categoryRepo:   categoryRepo,
	}
}

// Execute executes the listing of expenses.
func (uc *ListExpensesUseCase) Execute(ctx context.Context, input ListExpensesInput) (*ListExpensesOutput, error) {
	// 1. Check permissions
	perms, err := uc.permissionRepo.GetByUserID(ctx, input.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	hasPermission := false
	for _, p := range perms {
		if p.Code == "expenses.read" || p.Code == "system.all" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: expenses.read", nil)
	}

	// 2. Prepare filter
	filter := repository.ExpenseFilter{
		Limit:         input.Limit,
		Offset:        input.Offset,
		CategoryID:    input.CategoryID,
		ContractorID:  input.ContractorID,
		PaymentStatus: input.PaymentStatus,
		IsApproved:    input.IsApproved,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		SearchQuery:   input.SearchQuery,
		OrderBy:       "expense_date",
		OrderDesc:     true,
	}

	// 3. Get data
	expenses, err := uc.expenseRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}

	count, err := uc.expenseRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count expenses: %w", err)
	}

	// 4. Convert
	outputs := make([]*ExpenseOutput, len(expenses))
	for i, e := range expenses {
		outputs[i] = toOutput(e, uc.categoryRepo)
	}

	return &ListExpensesOutput{
		Expenses: outputs,
		Total:    count,
	}, nil
}
