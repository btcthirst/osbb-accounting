// application/service/cashflow_service.go
package service

import (
	"context"
	"fmt"
	"strings"

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
	cashFlowUseCase *cashflow.CashFlowUseCase
	paymentRepo     repository.PaymentRepository
	expenseRepo     repository.ExpenseRepository
	ownershipRepo   repository.OwnershipShareRepository
	apartmentRepo   repository.ApartmentRepository
	ownerRepo       repository.OwnerRepository
	categoryRepo    repository.ExpenseCategoryRepository
	contractorRepo  repository.ContractorRepository
}

// NewCashFlowService створює новий сервіс
func NewCashFlowService(
	cashFlowUseCase *cashflow.CashFlowUseCase,
	paymentRepo repository.PaymentRepository,
	expenseRepo repository.ExpenseRepository,
	ownershipRepo repository.OwnershipShareRepository,
	apartmentRepo repository.ApartmentRepository,
	ownerRepo repository.OwnerRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	contractorRepo repository.ContractorRepository,
) *CashFlowService {
	return &CashFlowService{
		cashFlowUseCase: cashFlowUseCase,
		paymentRepo:     paymentRepo,
		expenseRepo:     expenseRepo,
		ownershipRepo:   ownershipRepo,
		apartmentRepo:   apartmentRepo,
		ownerRepo:       ownerRepo,
		categoryRepo:    categoryRepo,
		contractorRepo:  contractorRepo,
	}
}

// List отримує список операцій руху коштів з додатковою інформацією.
func (s *CashFlowService) List(ctx context.Context, input cashflow.ListCashFlowInput) (*cashflow.ListCashFlowOutput, error) {
	output, err := s.cashFlowUseCase.List(ctx, input)
	if err != nil {
		return nil, err
	}

	// Заповнюємо додаткові дані для кожного запису
	if err := s.enrichEntries(ctx, output.Entries); err != nil {
		return nil, fmt.Errorf("failed to enrich entries: %w", err)
	}

	return output, nil
}

// enrichEntries заповнює додаткові дані (контрагенти, категорії) та розподіляє по колонках
func (s *CashFlowService) enrichEntries(ctx context.Context, entries []*cashflow.CashFlowEntry) error {
	// Кешуємо дані для мінімізації запитів до БД
	ownershipCache := make(map[int64]string)
	categoryCache := make(map[int64]string)
	contractorCache := make(map[int64]string)
	categoryCodeCache := make(map[int64]string)

	for _, entry := range entries {
		if entry.Type == "payment" {
			// Отримуємо інформацію про платника (власника)
			payment, err := s.paymentRepo.GetByID(ctx, entry.ID)
			if err == nil && payment != nil {
				entry.Counterparty = s.getOwnerLastName(ctx, payment.OwnershipShareID, ownershipCache)
			}
		} else if entry.Type == "contractor_payment" {
			// Отримуємо інформацію про контрагента
			// TODO: Додати логіку отримання назви контрагента
			entry.Counterparty = "Контрагент"
		} else if entry.Type == "expense" {
			// Отримуємо інформацію про витрату
			expense, err := s.expenseRepo.GetByID(ctx, entry.ID)
			if err == nil && expense != nil {
				// Отримуємо категорію
				entry.Category = s.getCategoryInfo(ctx, expense.CategoryID, categoryCache)
				categoryCode := s.getCategoryCode(ctx, expense.CategoryID, categoryCodeCache)
				entry.CategoryCode = categoryCode

				// Розподіляємо суму по колонках за кодом категорії
				switch categoryCode {
				case "313":
					entry.Credit313 = expense.Amount
				case "63":
					entry.Credit63 = expense.Amount
				case "641":
					entry.Credit641 = expense.Amount
				case "641.1":
					entry.Credit6411 = expense.Amount
				case "651":
					entry.Credit651 = expense.Amount
				case "94":
					entry.Credit94 = expense.Amount
				}

				// Контрагент (якщо є) або категорія
				if expense.ContractorID != nil {
					entry.Counterparty = s.getContractorInfo(ctx, expense.ContractorID, contractorCache)
				} else {
					// Для податків та ін.шов без контрагента - показуємо категорію
					entry.Counterparty = entry.Category
				}
			}
		}
	}

	return nil
}

// getOwnerLastName отримує прізвище власника (капіталізоване)
func (s *CashFlowService) getOwnerLastName(ctx context.Context, ownershipID int64, cache map[int64]string) string {
	if name, ok := cache[ownershipID]; ok {
		return name
	}

	ownership, err := s.ownershipRepo.GetByID(ctx, ownershipID)
	if err != nil || ownership == nil {
		cache[ownershipID] = "Невідомо"
		return "Невідомо"
	}

	owner, _ := s.ownerRepo.GetByID(ctx, ownership.OwnerID)
	var name string
	if owner != nil {
		// Капіталізуємо прізвище
		name = strings.ToUpper(owner.LastName)
	} else {
		name = "Невідомо"
	}

	cache[ownershipID] = name
	return name
}

// getCategoryInfo отримує назву категорії витрат
func (s *CashFlowService) getCategoryInfo(ctx context.Context, categoryID int64, cache map[int64]string) string {
	if name, ok := cache[categoryID]; ok {
		return name
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil || category == nil {
		cache[categoryID] = "Невідомо"
		return "Невідомо"
	}

	cache[categoryID] = category.Name
	return category.Name
}

// getCategoryCode отримує код категорії витрат
func (s *CashFlowService) getCategoryCode(ctx context.Context, categoryID int64, cache map[int64]string) string {
	if code, ok := cache[categoryID]; ok {
		return code
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil || category == nil || category.Code == nil {
		cache[categoryID] = ""
		return ""
	}

	cache[categoryID] = *category.Code
	return *category.Code
}

// getContractorInfo отримує назву контрагента
func (s *CashFlowService) getContractorInfo(ctx context.Context, contractorID *int64, cache map[int64]string) string {
	if contractorID == nil {
		return "-"
	}

	if name, ok := cache[*contractorID]; ok {
		return name
	}

	contractor, err := s.contractorRepo.GetByID(ctx, *contractorID)
	if err != nil || contractor == nil {
		cache[*contractorID] = "Невідомо"
		return "Невідомо"
	}

	cache[*contractorID] = contractor.Name
	return contractor.Name
}

// ExportToXLSX експортує дані руху коштів в XLSX файл.
func (s *CashFlowService) ExportToXLSX(ctx context.Context, input cashflow.ListCashFlowInput, osbbName string) (*excelize.File, string, error) {
	// Отримуємо дані
	output, err := s.List(ctx, input)
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

	file, err := cashflow.ExportToXLSX(exportInput)
	if err != nil {
		return nil, "", err
	}

	filename := cashflow.GenerateFilename(*input.Month, *input.Year)
	return file, filename, nil
}
