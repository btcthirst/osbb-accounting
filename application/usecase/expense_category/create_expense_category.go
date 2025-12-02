package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// CreateExpenseCategoryUseCase
// ============================================================================

type CreateExpenseCategoryUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewCreateExpenseCategoryUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *CreateExpenseCategoryUseCase {
	return &CreateExpenseCategoryUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type CreateExpenseCategoryInput struct {
	CurrentUserID int64
	Name          string
	Code          *string
	ParentID      *int64
	CategoryType  entity.CategoryType
	Description   *string
}

func (uc *CreateExpenseCategoryUseCase) Execute(
	ctx context.Context,
	input CreateExpenseCategoryInput,
) (*ExpenseCategoryOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "expenses", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"insufficient permissions to create expense category",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Перевірка унікальності коду
	if input.Code != nil && *input.Code != "" {
		exists, err := uc.categoryRepo.ExistsByCode(ctx, *input.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to check code: %w", err)
		}
		if exists {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"category with this code already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "code")
		}
	}

	// Перевірка батьківської категорії
	if input.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			if domainErrors.Is(err, domainErrors.ErrNotFound) {
				return nil, domainErrors.NewDomainError(
					domainErrors.CodeNotFound,
					"parent category not found",
					entity.ErrCategoryParentNotFound,
				)
			}
			return nil, fmt.Errorf("failed to get parent category: %w", err)
		}

		if !parent.IsActive {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeValidationFailed,
				"parent category is inactive",
				entity.ErrCategoryParentInactive,
			)
		}
	}

	// Створення entity
	category, err := entity.NewExpenseCategory(
		input.Name,
		input.CategoryType,
		input.ParentID,
		input.Code,
		input.Description,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid category data",
			err,
		)
	}

	// Збереження
	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}

	// Отримуємо з повною інформацією
	return uc.buildOutput(ctx, category)
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *CreateExpenseCategoryUseCase) buildOutput(
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
