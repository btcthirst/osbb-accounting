package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// UpdateExpenseCategoryUseCase
// ============================================================================

type UpdateExpenseCategoryUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdateExpenseCategoryUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateExpenseCategoryUseCase {
	return &UpdateExpenseCategoryUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdateExpenseCategoryInput struct {
	CurrentUserID int64
	CategoryID    int64
	Name          string
	CategoryType  entity.CategoryType
	Description   *string
}

func (uc *UpdateExpenseCategoryUseCase) Execute(
	ctx context.Context,
	input UpdateExpenseCategoryInput,
) (*ExpenseCategoryOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "expenses", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання поточних даних
	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"expense category not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}

	// Оновлення
	if err := category.Update(
		input.Name,
		input.CategoryType,
		input.Description,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid category data",
			err,
		)
	}

	// Збереження
	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update expense category: %w", err)
	}

	return uc.buildOutput(ctx, category)
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *UpdateExpenseCategoryUseCase) buildOutput(
	ctx context.Context,
	category *entity.ExpenseCategory,
) (*ExpenseCategoryOutput, error) {
	hasChildren, err := uc.categoryRepo.HasChildren(ctx, category.ID)
	if err != nil {
		hasChildren = false
	}

	return &ExpenseCategoryOutput{
		ID:              category.ID,
		Name:            category.Name,
		Code:            category.Code,
		ParentID:        category.ParentID,
		CategoryType:    category.CategoryType,
		TypeDisplayName: category.CategoryType.GetDisplayName(),
		Description:     category.Description,
		IsActive:        category.IsActive,
		Level:           category.Level,
		Path:            category.GetFullPath(),
		DisplayName:     category.GetDisplayName(),
		HasChildren:     hasChildren,
		CreatedAt:       category.CreatedAt,
		UpdatedAt:       category.UpdatedAt,
	}, nil
}
