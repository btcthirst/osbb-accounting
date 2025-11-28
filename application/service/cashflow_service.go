// application/service/cashflow_service.go
package service

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/cashflow"
	"osbb-accounting/domain/repository"
)

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

// List отримує список операцій руху коштів з додатковою інформацією
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

// enrichEntries заповнює додаткові дані (контрагенти, категорії)
func (s *CashFlowService) enrichEntries(ctx context.Context, entries []*cashflow.CashFlowEntry) error {
	// Кешуємо дані для мінімізації запитів до БД
	ownershipCache := make(map[int64]string)
	categoryCache := make(map[int64]string)
	contractorCache := make(map[int64]string)

	for _, entry := range entries {
		if entry.Type == "payment" {
			// Отримуємо інформацію про платника
			payment, err := s.paymentRepo.GetByID(ctx, entry.ID)
			if err == nil && payment != nil {
				entry.Counterparty = s.getOwnershipInfo(ctx, payment.OwnershipShareID, ownershipCache)
			}
		} else if entry.Type == "expense" {
			// Отримуємо інформацію про витрату
			expense, err := s.expenseRepo.GetByID(ctx, entry.ID)
			if err == nil && expense != nil {
				entry.Counterparty = s.getContractorInfo(ctx, expense.ContractorID, contractorCache)
				entry.Category = s.getCategoryInfo(ctx, expense.CategoryID, categoryCache)
			}
		}
	}

	return nil
}

// getOwnershipInfo отримує інформацію про власника
func (s *CashFlowService) getOwnershipInfo(ctx context.Context, ownershipID int64, cache map[int64]string) string {
	if name, ok := cache[ownershipID]; ok {
		return name
	}

	ownership, err := s.ownershipRepo.GetByID(ctx, ownershipID)
	if err != nil || ownership == nil {
		cache[ownershipID] = "Невідомо"
		return "Невідомо"
	}

	// Отримуємо інформацію про квартиру та власника
	apartment, _ := s.apartmentRepo.GetByID(ctx, ownership.ApartmentID)
	owner, _ := s.ownerRepo.GetByID(ctx, ownership.OwnerID)

	var name string
	if apartment != nil && owner != nil {
		name = fmt.Sprintf("%s %s (Кв. %s)", owner.FirstName, owner.LastName, apartment.ApartmentNumber)
	} else if apartment != nil {
		name = fmt.Sprintf("Кв. %s", apartment.ApartmentNumber)
	} else if owner != nil {
		name = fmt.Sprintf("%s %s", owner.FirstName, owner.LastName)
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
