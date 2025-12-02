package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UnapproveExpenseUseCase unapproves an expense.
type UnapproveExpenseUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
}

// NewUnapproveExpenseUseCase creates a new use case.
func NewUnapproveExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
) *UnapproveExpenseUseCase {
	return &UnapproveExpenseUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
	}
}

// Execute executes the unapproval of an expense.
func (uc *UnapproveExpenseUseCase) Execute(ctx context.Context, input UnapproveExpenseInput) (bool, error) {
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

	if err := e.Unapprove(); err != nil {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, err.Error(), err)
	}

	if err := uc.expenseRepo.Update(ctx, e); err != nil {
		return false, err
	}

	return true, nil
}
