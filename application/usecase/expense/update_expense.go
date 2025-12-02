package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateExpenseUseCase updates an existing expense.
type UpdateExpenseUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
	categoryRepo   repository.ExpenseCategoryRepository
}

// NewUpdateExpenseUseCase creates a new use case.
func NewUpdateExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
		categoryRepo:   categoryRepo,
	}
}

// Execute executes the update of an expense.
func (uc *UpdateExpenseUseCase) Execute(ctx context.Context, input UpdateExpenseInput) (*ExpenseOutput, error) {
	// 1. Check permissions
	perms, err := uc.permissionRepo.GetByUserID(ctx, input.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	hasPermission := false
	for _, p := range perms {
		if p.Code == "expenses.update" || p.Code == "system.all" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: expenses.update", nil)
	}

	// 2. Get existing
	e, err := uc.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	// 3. Check if approved
	if e.IsApproved() {
		return nil, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, "cannot update approved expense", nil)
	}

	// 4. Update
	if err := e.Update(
		input.CategoryID,
		input.ContractorID,
		input.ExpenseDate,
		input.Amount,
		input.Description,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
	}

	// 5. Save
	if err := uc.expenseRepo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	return toOutput(e, uc.categoryRepo), nil
}
