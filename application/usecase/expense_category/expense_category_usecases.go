// application/usecase/expense_category/expense_category_usecases.go
package expense_category

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// OUTPUT TYPES
// ============================================================================

// ExpenseCategoryOutput - результат операцій з категорією витрат.
type ExpenseCategoryOutput struct {
	ID              int64
	Name            string
	Code            *string
	ParentID        *int64
	CategoryType    entity.CategoryType
	TypeDisplayName string
	Description     *string
	IsActive        bool
	Level           int
	Path            string
	DisplayName     string
	HasChildren     bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ExpenseCategoryTreeOutput - категорія з дочірніми елементами.
type ExpenseCategoryTreeOutput struct {
	ExpenseCategoryOutput
	Children []*ExpenseCategoryTreeOutput
}

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
// GetExpenseCategoryUseCase
// ============================================================================

type GetExpenseCategoryUseCase struct {
	categoryRepo   repository.ExpenseCategoryRepository
	permissionRepo repository.PermissionRepository
}

func NewGetExpenseCategoryUseCase(
	categoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *GetExpenseCategoryUseCase {
	return &GetExpenseCategoryUseCase{
		categoryRepo:   categoryRepo,
		permissionRepo: permissionRepo,
	}
}

type GetExpenseCategoryInput struct {
	CurrentUserID int64
	CategoryID    int64
}

func (uc *GetExpenseCategoryUseCase) Execute(
	ctx context.Context,
	input GetExpenseCategoryInput,
) (*ExpenseCategoryOutput, error) {
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

	// Отримання категорії
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

	return uc.buildOutput(ctx, category)
}

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

type DeleteExpenseCategoryOutput struct {
	Success bool
	Message string
}

func (uc *DeleteExpenseCategoryUseCase) Execute(
	ctx context.Context,
	input DeleteExpenseCategoryInput,
) (*DeleteExpenseCategoryOutput, error) {
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

	return &DeleteExpenseCategoryOutput{
		Success: true,
		Message: fmt.Sprintf("Category '%s' successfully deleted", category.GetDisplayName()),
	}, nil
}

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

func (uc *GetExpenseCategoryUseCase) buildOutput(
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
