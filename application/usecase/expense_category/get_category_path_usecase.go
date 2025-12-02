package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetCategoryPathUseCase - отримання шляху до категорії
// ============================================================================

type GetCategoryPathUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewGetCategoryPathUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *GetCategoryPathUseCase {
	return &GetCategoryPathUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type GetCategoryPathInput struct {
	CurrentUserID int64
	CategoryID    int64
}

type CategoryPathOutput struct {
	Path       []*ExpenseCategoryOutput
	FullPath   string
	Depth      int
	IsRoot     bool
	RootID     int64
	RootName   string
	ParentID   *int64
	ParentName *string
}

func (uc *GetCategoryPathUseCase) Execute(
	ctx context.Context,
	input GetCategoryPathInput,
) (*CategoryPathOutput, error) {
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

	// Отримання поточної категорії
	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Отримання предків
	ancestors, err := uc.categoryRepo.GetAncestors(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}

	// Будуємо шлях (від кореня до поточної)
	path := make([]*ExpenseCategoryOutput, 0, len(ancestors)+1)
	for i := len(ancestors) - 1; i >= 0; i-- {
		output, err := uc.buildOutput(ctx, ancestors[i])
		if err != nil {
			return nil, err
		}
		path = append(path, output)
	}

	// Додаємо поточну категорію
	currentOutput, err := uc.buildOutput(ctx, category)
	if err != nil {
		return nil, err
	}
	path = append(path, currentOutput)

	// Будуємо повний шлях
	fullPath := ""
	for i, p := range path {
		if i > 0 {
			fullPath += " / "
		}
		fullPath += p.Name
	}

	// Визначаємо root та parent
	output := &CategoryPathOutput{
		Path:     path,
		FullPath: fullPath,
		Depth:    len(path) - 1,
		IsRoot:   category.ParentID == nil,
	}

	if len(path) > 0 {
		output.RootID = path[0].ID
		output.RootName = path[0].Name
	}

	if category.ParentID != nil && len(path) > 1 {
		parent := path[len(path)-2]
		output.ParentID = &parent.ID
		output.ParentName = &parent.Name
	}

	return output, nil
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *GetCategoryPathUseCase) buildOutput(
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
