package expense

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetExpenseStatisticsUseCase gets expense statistics.
type GetExpenseStatisticsUseCase struct {
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
}

// NewGetExpenseStatisticsUseCase creates a new use case.
func NewGetExpenseStatisticsUseCase(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
) *GetExpenseStatisticsUseCase {
	return &GetExpenseStatisticsUseCase{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
	}
}

// Execute executes the retrieval of expense statistics.
func (uc *GetExpenseStatisticsUseCase) Execute(ctx context.Context, input GetExpenseStatisticsInput) (*repository.ExpenseStatistics, error) {
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

	filter := repository.ExpenseFilter{
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	return uc.expenseRepo.GetStatistics(ctx, filter)
}
