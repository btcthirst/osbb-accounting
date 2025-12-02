package expense_category

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// ListExpenseCategoriesUseCase
// ============================================================================

type ListExpenseCategoriesUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewListExpenseCategoriesUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *ListExpenseCategoriesUseCase {
	return &ListExpenseCategoriesUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type ListExpenseCategoriesInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string
	CategoryType   *entity.CategoryType
	ParentID       *int64
	OnlyRoot       bool
	MaxDepth       *int
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

type ListExpenseCategoriesOutput struct {
	Categories []*ExpenseCategoryOutput
	Total      int64
	Limit      int
	Offset     int
	HasMore    bool
}

func (uc *ListExpenseCategoriesUseCase) Execute(
	ctx context.Context,
	input ListExpenseCategoriesInput,
) (*ListExpenseCategoriesOutput, error) {
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

	// Підготовка фільтру
	filter := repository.ExpenseCategoryFilter{
		IncludeDeleted: input.IncludeDeleted,
		IsActive:       input.IsActive,
		SearchQuery:    input.SearchQuery,
		CategoryType:   input.CategoryType,
		ParentID:       input.ParentID,
		OnlyRoot:       input.OnlyRoot,
		MaxDepth:       input.MaxDepth,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку
	categories, err := uc.categoryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list expense categories: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.categoryRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count expense categories: %w", err)
	}

	// Конвертація в output
	outputs := make([]*ExpenseCategoryOutput, len(categories))
	for i, cat := range categories {
		output, err := uc.buildOutput(ctx, cat)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	return &ListExpenseCategoriesOutput{
		Categories: outputs,
		Total:      total,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		HasMore:    int64(filter.Offset+filter.Limit) < total,
	}, nil
}

// ============================================================================
// Helper функції
// ============================================================================

func (uc *ListExpenseCategoriesUseCase) buildOutput(
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
