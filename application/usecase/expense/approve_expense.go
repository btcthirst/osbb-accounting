package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ApproveExpenseUseCase approves an expense.
type ApproveExpenseUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
}

// NewApproveExpenseUseCase creates a new use case.
func NewApproveExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
) *ApproveExpenseUseCase {
	return &ApproveExpenseUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
	}
}

// Execute executes the approval of an expense.
func (uc *ApproveExpenseUseCase) Execute(ctx context.Context, input ApproveExpenseInput) (bool, error) {
	perms, err := uc.permissionRepo.GetByUserID(ctx, input.CurrentUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get permissions: %w", err)
	}

	hasPermission := false
	for _, p := range perms {
		if p.Code == "expenses.approve" || p.Code == "system.all" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return false, domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: expenses.approve", nil)
	}

	e, err := uc.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return false, err
	}

	if err := e.Approve(input.CurrentUserID); err != nil {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, err.Error(), err)
	}

	if err := uc.expenseRepo.Update(ctx, e); err != nil {
		return false, err
	}

	return true, nil
}
