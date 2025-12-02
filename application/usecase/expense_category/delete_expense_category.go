package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// DeleteExpenseCategoryUseCase
// ============================================================================

type DeleteExpenseCategoryUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewDeleteExpenseCategoryUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteExpenseCategoryUseCase {
	return &DeleteExpenseCategoryUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteExpenseCategoryInput struct {
	CurrentUserID int64
	CategoryID    int64
}

func (uc *DeleteExpenseCategoryUseCase) Execute(
	ctx context.Context,
	input DeleteExpenseCategoryInput,
) (*shared.DeleteOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "expenses", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання для інформації
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

	// Перевірка наявності дочірніх категорій
	hasChildren, err := uc.categoryRepo.HasChildren(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check children: %w", err)
	}
	if hasChildren {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"cannot delete category with children",
			entity.ErrCategoryHasChildren,
		)
	}

	// Видалення
	if err := uc.categoryRepo.SoftDelete(ctx, input.CategoryID); err != nil {
		return nil, fmt.Errorf("failed to delete expense category: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Category '%s' successfully deleted", category.GetDisplayName()),
	), nil
}
