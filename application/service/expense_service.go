// application/service/expense_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/expense"
	"osbb-accounting/domain/repository"
)

// ExpenseServiceInterface defines the contract for expense service
type ExpenseServiceInterface interface {
	Create(ctx context.Context, input expense.CreateExpenseInput) (*expense.ExpenseOutput, error)
	Update(ctx context.Context, input expense.UpdateExpenseInput) (*expense.ExpenseOutput, error)
	Delete(ctx context.Context, input expense.DeleteExpenseInput) (bool, error)
	Get(ctx context.Context, id int64) (*expense.ExpenseOutput, error)
	List(ctx context.Context, input expense.ListExpensesInput) (*expense.ListExpensesOutput, error)
	Approve(ctx context.Context, input expense.ApproveExpenseInput) (bool, error)
	Unapprove(ctx context.Context, input expense.UnapproveExpenseInput) (bool, error)
	GetStatistics(ctx context.Context, input expense.GetExpenseStatisticsInput) (*repository.ExpenseStatistics, error)
}

// ExpenseService implements ExpenseServiceInterface
type ExpenseService struct {
	createUseCase        *expense.CreateExpenseUseCase
	updateUseCase        *expense.UpdateExpenseUseCase
	deleteUseCase        *expense.DeleteExpenseUseCase
	getUseCase           *expense.GetExpenseUseCase
	listUseCase          *expense.ListExpensesUseCase
	approveUseCase       *expense.ApproveExpenseUseCase
	unapproveUseCase     *expense.UnapproveExpenseUseCase
	getStatisticsUseCase *expense.GetExpenseStatisticsUseCase
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *ExpenseService {
	return &ExpenseService{
		createUseCase:        expense.NewCreateExpenseUseCase(expenseRepo, permissionRepo, categoryRepo),
		updateUseCase:        expense.NewUpdateExpenseUseCase(expenseRepo, permissionRepo, categoryRepo),
		deleteUseCase:        expense.NewDeleteExpenseUseCase(expenseRepo, permissionRepo),
		getUseCase:           expense.NewGetExpenseUseCase(expenseRepo, categoryRepo),
		listUseCase:          expense.NewListExpensesUseCase(expenseRepo, permissionRepo, categoryRepo),
		approveUseCase:       expense.NewApproveExpenseUseCase(expenseRepo, permissionRepo),
		unapproveUseCase:     expense.NewUnapproveExpenseUseCase(expenseRepo, permissionRepo),
		getStatisticsUseCase: expense.NewGetExpenseStatisticsUseCase(expenseRepo, permissionRepo),
	}
}

// Create creates a new expense
func (s *ExpenseService) Create(ctx context.Context, input expense.CreateExpenseInput) (*expense.ExpenseOutput, error) {
	return s.createUseCase.Execute(ctx, input)
}

// Update updates an existing expense
func (s *ExpenseService) Update(ctx context.Context, input expense.UpdateExpenseInput) (*expense.ExpenseOutput, error) {
	return s.updateUseCase.Execute(ctx, input)
}

// Delete deletes an expense
func (s *ExpenseService) Delete(ctx context.Context, input expense.DeleteExpenseInput) (bool, error) {
	return s.deleteUseCase.Execute(ctx, input)
}

// Get gets an expense by ID
func (s *ExpenseService) Get(ctx context.Context, id int64) (*expense.ExpenseOutput, error) {
	return s.getUseCase.Execute(ctx, id)
}

// List lists expenses
func (s *ExpenseService) List(ctx context.Context, input expense.ListExpensesInput) (*expense.ListExpensesOutput, error) {
	return s.listUseCase.Execute(ctx, input)
}

// Approve approves an expense
func (s *ExpenseService) Approve(ctx context.Context, input expense.ApproveExpenseInput) (bool, error) {
	return s.approveUseCase.Execute(ctx, input)
}

// Unapprove unapproves an expense
func (s *ExpenseService) Unapprove(ctx context.Context, input expense.UnapproveExpenseInput) (bool, error) {
	return s.unapproveUseCase.Execute(ctx, input)
}

// GetStatistics gets expense statistics
func (s *ExpenseService) GetStatistics(ctx context.Context, input expense.GetExpenseStatisticsInput) (*repository.ExpenseStatistics, error) {
	return s.getStatisticsUseCase.Execute(ctx, input)
}
