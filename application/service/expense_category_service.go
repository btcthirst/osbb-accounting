package service

import (
	"context"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/repository"
)

type ExpenseCategoryServiceInterface interface {
	Create(ctx context.Context, input expense_category.CreateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	Update(ctx context.Context, input expense_category.UpdateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	Delete(ctx context.Context, input expense_category.DeleteExpenseCategoryInput) (*shared.DeleteOutput, error)
	GetCategoryPath(ctx context.Context, input expense_category.GetCategoryPathInput) (*expense_category.CategoryPathOutput, error)
	GetCategoryTree(ctx context.Context, input expense_category.GetCategoryTreeInput) ([]*expense_category.ExpenseCategoryTreeOutput, error)
	Get(ctx context.Context, input expense_category.GetExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	List(ctx context.Context, input expense_category.ListExpenseCategoriesInput) (*expense_category.ListExpenseCategoriesOutput, error)
	MoveCategory(ctx context.Context, input expense_category.MoveCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
}

type ExpenseCategoryService struct {
	createExpenseCategory *expense_category.CreateExpenseCategoryUseCase
	updateExpenseCategory *expense_category.UpdateExpenseCategoryUseCase
	deleteExpenseCategory *expense_category.DeleteExpenseCategoryUseCase
	getCategoryPath       *expense_category.GetCategoryPathUseCase
	getCategoryTree       *expense_category.GetCategoryTreeUseCase
	getExpenseCategory    *expense_category.GetExpenseCategoryUseCase
	listExpenseCategories *expense_category.ListExpenseCategoriesUseCase
	moveCategory          *expense_category.MoveCategoryUseCase
}

func NewExpenseCategoryService(
	expenseCategoryRepo repository.ExpenseCategoryRepository,
	permissionRepo repository.PermissionRepository,
) *ExpenseCategoryService {
	return &ExpenseCategoryService{
		createExpenseCategory: expense_category.NewCreateExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		updateExpenseCategory: expense_category.NewUpdateExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		deleteExpenseCategory: expense_category.NewDeleteExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		getCategoryPath:       expense_category.NewGetCategoryPathUseCase(expenseCategoryRepo, permissionRepo),
		getCategoryTree:       expense_category.NewGetCategoryTreeUseCase(expenseCategoryRepo, permissionRepo),
		getExpenseCategory:    expense_category.NewGetExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		listExpenseCategories: expense_category.NewListExpenseCategoriesUseCase(expenseCategoryRepo, permissionRepo),
		moveCategory:          expense_category.NewMoveCategoryUseCase(expenseCategoryRepo, permissionRepo),
	}
}

// Create створює нову категорію витрат.
func (s *ExpenseCategoryService) Create(ctx context.Context, input expense_category.CreateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.createExpenseCategory.Execute(ctx, input)
}

// Update оновлює дані категорії витрат.
func (s *ExpenseCategoryService) Update(ctx context.Context, input expense_category.UpdateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.updateExpenseCategory.Execute(ctx, input)
}

// Delete видаляє категорію витрат.
func (s *ExpenseCategoryService) Delete(ctx context.Context, input expense_category.DeleteExpenseCategoryInput) (*shared.DeleteOutput, error) {
	return s.deleteExpenseCategory.Execute(ctx, input)
}

// GetCategoryPath отримує шлях категорії (від кореня).
func (s *ExpenseCategoryService) GetCategoryPath(ctx context.Context, input expense_category.GetCategoryPathInput) (*expense_category.CategoryPathOutput, error) {
	return s.getCategoryPath.Execute(ctx, input)
}

// GetCategoryTree отримує дерево категорій.
func (s *ExpenseCategoryService) GetCategoryTree(ctx context.Context, input expense_category.GetCategoryTreeInput) ([]*expense_category.ExpenseCategoryTreeOutput, error) {
	return s.getCategoryTree.Execute(ctx, input)
}

// Get отримує категорію витрат за ID.
func (s *ExpenseCategoryService) Get(ctx context.Context, input expense_category.GetExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.getExpenseCategory.Execute(ctx, input)
}

// List отримує список категорій витрат з фільтрацією.
func (s *ExpenseCategoryService) List(ctx context.Context, input expense_category.ListExpenseCategoriesInput) (*expense_category.ListExpenseCategoriesOutput, error) {
	return s.listExpenseCategories.Execute(ctx, input)
}

// MoveCategory переміщує категорію до іншого батьківського вузла.
func (s *ExpenseCategoryService) MoveCategory(ctx context.Context, input expense_category.MoveCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.moveCategory.Execute(ctx, input)
}
