// application/service/cashflow_service.go
package service

import (
	"context"

	"github.com/xuri/excelize/v2"

	"osbb-accounting/application/usecase/cashflow"
	"osbb-accounting/domain/repository"
)

// CashFlowServiceInterface defines the contract for cash flow operations.
type CashFlowServiceInterface interface {
	List(ctx context.Context, input cashflow.ListCashFlowInput) (*cashflow.ListCashFlowOutput, error)
	ExportToXLSX(ctx context.Context, input cashflow.ListCashFlowInput, osbbName string) (*excelize.File, string, error)
}

// CashFlowService - сервісний шар для роботи з рухом коштів
type CashFlowService struct {
	listUseCase   *cashflow.ListCashFlowUseCase
	exportUseCase *cashflow.ExportCashFlowUseCase
}

// NewCashFlowService створює новий сервіс
func NewCashFlowService(
	paymentRepo repository.PaymentRepository,
	contractorPaymentRepo repository.ContractorPaymentRepository,
	expenseRepo repository.ExpenseRepository,
	ownershipRepo repository.OwnershipShareRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	contractorRepo repository.ContractorRepository,
	ownerRepo repository.OwnerRepository,
) *CashFlowService {
	return &CashFlowService{
		listUseCase: cashflow.NewListCashFlowUseCase(
			paymentRepo,
			contractorPaymentRepo,
			expenseRepo,
			ownershipRepo,
			categoryRepo,
			contractorRepo,
			ownerRepo,
		),
		exportUseCase: cashflow.NewExportCashFlowUseCase(),
	}
}

// List отримує список операцій руху коштів з додатковою інформацією.
func (s *CashFlowService) List(ctx context.Context, input cashflow.ListCashFlowInput) (*cashflow.ListCashFlowOutput, error) {
	return s.listUseCase.Execute(ctx, input)
}

// ExportToXLSX експортує дані руху коштів в XLSX файл.
func (s *CashFlowService) ExportToXLSX(ctx context.Context, input cashflow.ListCashFlowInput, osbbName string) (*excelize.File, string, error) {
	// Отримуємо дані
	output, err := s.listUseCase.Execute(ctx, input)
	if err != nil {
		return nil, "", err
	}

	// Генеруємо XLSX
	exportInput := cashflow.ExportToXLSXInput{
		Entries:      output.Entries,
		Month:        *input.Month,
		Year:         *input.Year,
		OSBBName:     osbbName,
		TotalDebit:   output.TotalDebit,
		TotalCredit:  output.TotalCredit,
		StartBalance: output.StartBalance,
		EndBalance:   output.EndBalance,
	}

	file, err := s.exportUseCase.Execute(exportInput)
	if err != nil {
		return nil, "", err
	}

	filename := cashflow.GenerateFilename(*input.Month, *input.Year)
	return file, filename, nil
}
