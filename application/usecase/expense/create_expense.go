package expense

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateExpenseUseCase creates a new expense.
type CreateExpenseUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
	categoryRepo   repository.ExpenseCategoryRepository
}

// NewCreateExpenseUseCase creates a new use case.
func NewCreateExpenseUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *CreateExpenseUseCase {
	return &CreateExpenseUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
		categoryRepo:   categoryRepo,
	}
}

// Execute executes the creation of an expense.
func (uc *CreateExpenseUseCase) Execute(ctx context.Context, input CreateExpenseInput) (*ExpenseOutput, error) {
	// 1. Check permissions
	// Note: We are using a simplified permission check here similar to what was in the service.
	// Ideally, this should be consistent with other use cases using permissionRepo.
	perms, err := uc.permissionRepo.GetByUserID(ctx, input.CurrentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	hasPermission := false
	for _, p := range perms {
		if p.Code == "expenses.create" || p.Code == "system.all" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: expenses.create", nil)
	}

	// 2. Create entity
	e, err := entity.NewExpense(
		input.CategoryID,
		input.ContractorID,
		input.ExpenseDate,
		input.Amount,
		input.Description,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
	}

	// 3. Set optional fields
	if input.DocumentType != nil && input.DocumentNumber != nil && input.DocumentDate != nil {
		if err := e.SetDocument(*input.DocumentType, *input.DocumentNumber, *input.DocumentDate); err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
		}
	}

	e.Notes = input.Notes

	// 4. Save
	if err := uc.expenseRepo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return toOutput(e, uc.categoryRepo), nil
}
