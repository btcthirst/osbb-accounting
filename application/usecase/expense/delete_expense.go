package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// DeleteExpenseUseCase deletes an expense.
type DeleteExpenseUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
}

// NewDeleteExpenseUseCase creates a new use case.
func NewDeleteExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteExpenseUseCase {
	return &DeleteExpenseUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
	}
}

// Execute executes the deletion of an expense.
func (uc *DeleteExpenseUseCase) Execute(ctx context.Context, input DeleteExpenseInput) (bool, error) {
	// 1. Check permissions
	perms, err := uc.permissionRepo.GetByUserID(ctx, input.CurrentUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get permissions: %w", err)
	}

	hasPermission := false
	for _, p := range perms {
		if p.Code == "expenses.delete" || p.Code == "system.all" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return false, domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: expenses.delete", nil)
	}

	// 2. Get existing
	e, err := uc.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return false, fmt.Errorf("failed to get expense: %w", err)
	}

	// 3. Check if approved
	if e.IsApproved() {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, "cannot delete approved expense", nil)
	}

	// 4. Delete
	if err := uc.expenseRepo.SoftDelete(ctx, input.ExpenseID); err != nil {
		return false, fmt.Errorf("failed to delete expense: %w", err)
	}

	return true, nil
}
