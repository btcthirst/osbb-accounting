package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetCategoryTreeUseCase - отримання дерева категорій
// ============================================================================

type GetCategoryTreeUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewGetCategoryTreeUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *GetCategoryTreeUseCase {
	return &GetCategoryTreeUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type GetCategoryTreeInput struct {
	CurrentUserID int64
}

func (uc *GetCategoryTreeUseCase) Execute(
	ctx context.Context,
	input GetCategoryTreeInput,
) ([]*ExpenseCategoryTreeOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "expenses", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання дерева
	tree, err := uc.categoryRepo.GetTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}

	// Конвертація в output
	return uc.buildTreeOutput(ctx, tree)
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *GetCategoryTreeUseCase) buildTreeOutput(
	ctx context.Context,
	categories []*entity.ExpenseCategory,
) ([]*ExpenseCategoryTreeOutput, error) {
	outputs := make([]*ExpenseCategoryTreeOutput, len(categories))

	for i, cat := range categories {
		output := &ExpenseCategoryTreeOutput{
			ExpenseCategoryOutput: ExpenseCategoryOutput{
				ID:              cat.ID,
				Name:            cat.Name,
				Code:            cat.Code,
				ParentID:        cat.ParentID,
				CategoryType:    cat.CategoryType,
				TypeDisplayName: cat.CategoryType.GetDisplayName(),
				Description:     cat.Description,
				IsActive:        cat.IsActive,
				Level:           cat.Level,
				Path:            cat.GetFullPath(),
				DisplayName:     cat.GetDisplayName(),
				HasChildren:     len(cat.Children) > 0,
				CreatedAt:       cat.CreatedAt,
				UpdatedAt:       cat.UpdatedAt,
			},
		}

		// Рекурсивно обробляємо дочірні елементи
		if len(cat.Children) > 0 {
			children, err := uc.buildTreeOutput(ctx, cat.Children)
			if err != nil {
				return nil, err
			}
			output.Children = children
		}

		outputs[i] = output
	}

	return outputs, nil
}
