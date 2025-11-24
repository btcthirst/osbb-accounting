// application/service/expense_service.go
package service

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/expense"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
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
	expenseRepo    repository.ExpenseRepository
	permissionRepo repository.PermissionRepository
	// categoryRepo repository.ExpenseCategoryRepository // Needed for category name
	// contractorRepo repository.ContractorRepository // Needed for contractor name
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	permissionRepo repository.PermissionRepository,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo:    expenseRepo,
		permissionRepo: permissionRepo,
	}
}

// Create creates a new expense
func (s *ExpenseService) Create(ctx context.Context, input expense.CreateExpenseInput) (*expense.ExpenseOutput, error) {
	// 1. Check permissions
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.create"); err != nil {
		return nil, err
	}

	// 2. Create entity
	e, err := entity.NewExpense(
		input.CategoryID,
		input.ContractorID,
		input.ExpenseDate,
		input.Amount,
		input.Description,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
	}

	// 3. Set optional fields
	if input.DocumentType != nil && input.DocumentNumber != nil && input.DocumentDate != nil {
		if err := e.SetDocument(*input.DocumentType, *input.DocumentNumber, *input.DocumentDate); err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
		}
	}

	e.Notes = input.Notes

	// 4. Save
	if err := s.expenseRepo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return s.toOutput(e), nil
}

// Update updates an existing expense
func (s *ExpenseService) Update(ctx context.Context, input expense.UpdateExpenseInput) (*expense.ExpenseOutput, error) {
	// 1. Check permissions
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.update"); err != nil {
		return nil, err
	}

	// 2. Get existing
	e, err := s.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	// 3. Check if approved
	if e.IsApproved() {
		return nil, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, "cannot update approved expense", nil)
	}

	// 4. Update
	if err := e.Update(
		input.CategoryID,
		input.ContractorID,
		input.ExpenseDate,
		input.Amount,
		input.Description,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.CodeValidationFailed, err.Error(), err)
	}

	// 5. Save
	if err := s.expenseRepo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	return s.toOutput(e), nil
}

// Delete deletes an expense
func (s *ExpenseService) Delete(ctx context.Context, input expense.DeleteExpenseInput) (bool, error) {
	// 1. Check permissions
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.delete"); err != nil {
		return false, err
	}

	// 2. Get existing
	e, err := s.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return false, fmt.Errorf("failed to get expense: %w", err)
	}

	// 3. Check if approved
	if e.IsApproved() {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, "cannot delete approved expense", nil)
	}

	// 4. Delete
	if err := s.expenseRepo.SoftDelete(ctx, input.ExpenseID); err != nil {
		return false, fmt.Errorf("failed to delete expense: %w", err)
	}

	return true, nil
}

// Get gets an expense by ID
func (s *ExpenseService) Get(ctx context.Context, id int64) (*expense.ExpenseOutput, error) {
	e, err := s.expenseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return s.toOutput(e), nil
}

// List lists expenses
func (s *ExpenseService) List(ctx context.Context, input expense.ListExpensesInput) (*expense.ListExpensesOutput, error) {
	// 1. Check permissions
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.read"); err != nil {
		return nil, err
	}

	// 2. Prepare filter
	filter := repository.ExpenseFilter{
		Limit:         input.Limit,
		Offset:        input.Offset,
		CategoryID:    input.CategoryID,
		ContractorID:  input.ContractorID,
		PaymentStatus: input.PaymentStatus,
		IsApproved:    input.IsApproved,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		SearchQuery:   input.SearchQuery,
		OrderBy:       "expense_date",
		OrderDesc:     true,
	}

	// 3. Get data
	expenses, err := s.expenseRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}

	count, err := s.expenseRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count expenses: %w", err)
	}

	// 4. Convert
	outputs := make([]*expense.ExpenseOutput, len(expenses))
	for i, e := range expenses {
		outputs[i] = s.toOutput(e)
	}

	return &expense.ListExpensesOutput{
		Expenses: outputs,
		Total:    count,
	}, nil
}

// Approve approves an expense
func (s *ExpenseService) Approve(ctx context.Context, input expense.ApproveExpenseInput) (bool, error) {
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.approve"); err != nil {
		return false, err
	}

	e, err := s.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return false, err
	}

	if err := e.Approve(input.CurrentUserID); err != nil {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, err.Error(), err)
	}

	if err := s.expenseRepo.Update(ctx, e); err != nil {
		return false, err
	}

	return true, nil
}

// Unapprove unapproves an expense
func (s *ExpenseService) Unapprove(ctx context.Context, input expense.UnapproveExpenseInput) (bool, error) {
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.approve"); err != nil {
		return false, err
	}

	e, err := s.expenseRepo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return false, err
	}

	if err := e.Unapprove(); err != nil {
		return false, domainErrors.NewDomainError(domainErrors.CodeOperationNotAllowed, err.Error(), err)
	}

	if err := s.expenseRepo.Update(ctx, e); err != nil {
		return false, err
	}

	return true, nil
}

// GetStatistics gets expense statistics
func (s *ExpenseService) GetStatistics(ctx context.Context, input expense.GetExpenseStatisticsInput) (*repository.ExpenseStatistics, error) {
	if err := s.checkPermission(ctx, input.CurrentUserID, "expenses.read"); err != nil {
		return nil, err
	}

	filter := repository.ExpenseFilter{
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	return s.expenseRepo.GetStatistics(ctx, filter)
}

// Helper: Check permission
func (s *ExpenseService) checkPermission(ctx context.Context, userID int64, permission string) error {
	// For simplicity, we assume permissionRepo has a method to check permissions or we use a separate auth service.
	// Here we just check if user has permission via repository if possible, or skip if not implemented.
	// In previous services we injected permissionRepo but didn't use it directly for check,
	// usually AuthManager handles UI checks, but service should also check.
	// Let's assume we trust the caller for now or implement a simple check if repo supports it.

	// Implementation detail: In this project, permissions are often checked at UI layer (AuthManager).
	// But for security, service layer should also check.
	// We will use permissionRepo to get user permissions and check.

	perms, err := s.permissionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	for _, p := range perms {
		if p.Code == permission || p.Code == "system.all" {
			return nil
		}
	}

	return domainErrors.NewDomainError(domainErrors.CodePermissionDenied, "permission denied: "+permission, nil)
}

// Helper: Convert to Output
func (s *ExpenseService) toOutput(e *entity.Expense) *expense.ExpenseOutput {
	return &expense.ExpenseOutput{
		ID:                e.ID,
		CategoryID:        e.CategoryID,
		CategoryName:      fmt.Sprintf("Category %d", e.CategoryID), // Placeholder, need CategoryRepo
		ContractorID:      e.ContractorID,
		ExpenseDate:       e.ExpenseDate,
		Amount:            e.Amount,
		Description:       e.Description,
		DocumentType:      e.DocumentType,
		DocumentNumber:    e.DocumentNumber,
		DocumentDate:      e.DocumentDate,
		PaymentStatus:     string(e.PaymentStatus),
		PaymentStatusName: e.PaymentStatus.GetDisplayName(),
		PaidAmount:        e.PaidAmount,
		PaymentDate:       e.PaymentDate,
		Notes:             e.Notes,
		IsApproved:        e.IsApproved(),
		ApprovedBy:        e.ApprovedBy,
		ApprovedAt:        e.ApprovedAt,
	}
}
