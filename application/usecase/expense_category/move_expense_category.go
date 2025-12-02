package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// MoveCategoryUseCase - переміщення категорії
// ============================================================================

type MoveCategoryUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewMoveCategoryUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *MoveCategoryUseCase {
	return &MoveCategoryUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type MoveCategoryInput struct {
	CurrentUserID int64
	CategoryID    int64
	NewParentID   *int64
}

func (uc *MoveCategoryUseCase) Execute(
	ctx context.Context,
	input MoveCategoryInput,
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

	// Переміщення
	if err := uc.categoryRepo.Move(ctx, input.CategoryID, input.NewParentID); err != nil {
		if domainErrors.Is(err, entity.ErrCategoryCircularReference) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeValidationFailed,
				"circular reference detected",
				err,
			)
		}
		if domainErrors.Is(err, entity.ErrCategoryMaxDepthExceeded) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeValidationFailed,
				"maximum depth exceeded",
				err,
			)
		}
		if domainErrors.Is(err, entity.ErrCategoryCannotBeOwnParent) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeValidationFailed,
				"category cannot be its own parent",
				err,
			)
		}
		return nil, fmt.Errorf("failed to move category: %w", err)
	}

	// Отримуємо оновлену категорію
	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated category: %w", err)
	}

	return uc.buildOutput(ctx, category)
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *MoveCategoryUseCase) buildOutput(
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
