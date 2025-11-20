package service

import (
	"context"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/domain/repository"
)

type ExpenseCategoryServiceI interface {
	Create(context.Context, expense_category.CreateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	Updete(context.Context, expense_category.UpdateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	Delete(context.Context, expense_category.DeleteExpenseCategoryInput) (*expense_category.DeleteExpenseCategoryOutput, error)
	GetCategoryPath(context.Context, expense_category.GetCategoryPathInput) (*expense_category.CategoryPathOutput, error)
	GetCategoryTree(context.Context, expense_category.GetCategoryTreeInput) ([]*expense_category.ExpenseCategoryTreeOutput, error)
	Get(context.Context, expense_category.GetExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
	List(context.Context, expense_category.ListExpenseCategoriesInput) (*expense_category.ListExpenseCategoriesOutput, error)
	MoveCategory(context.Context, expense_category.MoveCategoryInput) (*expense_category.ExpenseCategoryOutput, error)
}

type ExpenseCategoryService struct {
	createExpenseCategory *expense_category.CreateExpenseCategoryUseCase
	updateExpenseCategory *expense_category.UpdateExpenseCategoryUseCase
	deleteEpenseCategory  *expense_category.DeleteExpenseCategoryUseCase
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
		deleteEpenseCategory:  expense_category.NewDeleteExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		getCategoryPath:       expense_category.NewGetCategoryPathUseCase(expenseCategoryRepo, permissionRepo),
		getCategoryTree:       expense_category.NewGetCategoryTreeUseCase(expenseCategoryRepo, permissionRepo),
		getExpenseCategory:    expense_category.NewGetExpenseCategoryUseCase(expenseCategoryRepo, permissionRepo),
		listExpenseCategories: expense_category.NewListExpenseCategoriesUseCase(expenseCategoryRepo, permissionRepo),
		moveCategory:          expense_category.NewMoveCategoryUseCase(expenseCategoryRepo, permissionRepo),
	}
}

func (s *ExpenseCategoryService) Create(ctx context.Context, input expense_category.CreateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.createExpenseCategory.Execute(ctx, input)
}

func (s *ExpenseCategoryService) Updete(ctx context.Context, input expense_category.UpdateExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.updateExpenseCategory.Execute(ctx, input)
}

func (s *ExpenseCategoryService) Delete(ctx context.Context, input expense_category.DeleteExpenseCategoryInput) (*expense_category.DeleteExpenseCategoryOutput, error) {
	return s.deleteEpenseCategory.Execute(ctx, input)
}

func (s *ExpenseCategoryService) GetCategoryPath(ctx context.Context, input expense_category.GetCategoryPathInput) (*expense_category.CategoryPathOutput, error) {
	return s.getCategoryPath.Execute(ctx, input)
}

func (s *ExpenseCategoryService) GetCategoryTree(ctx context.Context, input expense_category.GetCategoryTreeInput) ([]*expense_category.ExpenseCategoryTreeOutput, error) {
	return s.getCategoryTree.Execute(ctx, input)
}

func (s *ExpenseCategoryService) Get(ctx context.Context, input expense_category.GetExpenseCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.getExpenseCategory.Execute(ctx, input)
}

func (s *ExpenseCategoryService) List(ctx context.Context, input expense_category.ListExpenseCategoriesInput) (*expense_category.ListExpenseCategoriesOutput, error) {
	return s.listExpenseCategories.Execute(ctx, input)
}

func (s *ExpenseCategoryService) MoveCategory(ctx context.Context, input expense_category.MoveCategoryInput) (*expense_category.ExpenseCategoryOutput, error) {
	return s.moveCategory.Execute(ctx, input)
}
