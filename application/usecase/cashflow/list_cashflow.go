// application/usecase/cashflow/list_cashflow.go
package cashflow

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"osbb-accounting/domain/repository"
)

// ListCashFlowUseCase - use case для отримання списку руху коштів
type ListCashFlowUseCase struct {
	paymentRepo           repository.PaymentRepository
	contractorPaymentRepo repository.ContractorPaymentRepository
	expenseRepo           repository.ExpenseRepository
	ownershipRepo         repository.OwnershipShareRepository
	categoryRepo          repository.ExpenseCategoryRepository
	contractorRepo        repository.ContractorRepository
	ownerRepo             repository.OwnerRepository
}

// NewListCashFlowUseCase створює новий use case
func NewListCashFlowUseCase(
	paymentRepo repository.PaymentRepository,
	contractorPaymentRepo repository.ContractorPaymentRepository,
	expenseRepo repository.ExpenseRepository,
	ownershipRepo repository.OwnershipShareRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	contractorRepo repository.ContractorRepository,
	ownerRepo repository.OwnerRepository,
) *ListCashFlowUseCase {
	return &ListCashFlowUseCase{
		paymentRepo:           paymentRepo,
		contractorPaymentRepo: contractorPaymentRepo,
		expenseRepo:           expenseRepo,
		ownershipRepo:         ownershipRepo,
		categoryRepo:          categoryRepo,
		contractorRepo:        contractorRepo,
		ownerRepo:             ownerRepo,
	}
}

// Execute отримує список операцій руху коштів
func (uc *ListCashFlowUseCase) Execute(ctx context.Context, input ListCashFlowInput) (*ListCashFlowOutput, error) {
	// Визначаємо діапазон дат
	var startDate, endDate time.Time
	if input.Month != nil && input.Year != nil {
		// Фільтр по місяцю/року
		startDate = time.Date(*input.Year, time.Month(*input.Month), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second) // Останній момент місяця
	} else if input.StartDate != nil && input.EndDate != nil {
		startDate = *input.StartDate
		endDate = *input.EndDate
	} else {
		// За замовчуванням - поточний місяць
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	}

	// Отримуємо платежі за період
	startUnix := startDate.Unix()
	endUnix := endDate.Unix()
	payments, err := uc.paymentRepo.List(ctx, repository.PaymentFilter{
		StartDate: &startUnix,
		EndDate:   &endUnix,
		Limit:     1000, // Достатньо для місячного звіту
	})
	if err != nil {
		return nil, err
	}

	// Отримуємо платежі від контрагентів за період
	contractorPayments, err := uc.contractorPaymentRepo.List(ctx, repository.ContractorPaymentFilter{
		StartDate: &startUnix,
		EndDate:   &endUnix,
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	// Отримуємо витрати за період
	expenses, err := uc.expenseRepo.List(ctx, repository.ExpenseFilter{
		StartDate: &startUnix,
		EndDate:   &endUnix,
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	// Збираємо всі записи в єдиний список
	entries := make([]*CashFlowEntry, 0, len(payments)+len(contractorPayments)+len(expenses))
	totalDebit := 0.0
	totalCredit := 0.0

	// Додаємо платежі від власників (надходження)
	for _, p := range payments {
		entry := &CashFlowEntry{
			ID:           p.ID,
			Date:         p.PaymentDate,
			Type:         "payment",
			Counterparty: "", // Буде заповнено в enrichEntries
			Description:  p.PaymentPurpose,
			Category:     "",
			CategoryCode: "",
			Debit:        p.Amount,
			Credit:       0,
		}
		entries = append(entries, entry)
		totalDebit += p.Amount
	}

	// Додаємо платежі від контрагентів (надходження)
	for _, cp := range contractorPayments {
		entry := &CashFlowEntry{
			ID:           cp.ID,
			Date:         cp.PaymentDate,
			Type:         "contractor_payment",
			Counterparty: "", // Буде заповнено в enrichEntries
			Description:  cp.Purpose,
			Category:     "",
			CategoryCode: "",
			Debit:        cp.Amount,
			Credit:       0,
		}
		entries = append(entries, entry)
		totalDebit += cp.Amount
	}

	// Додаємо витрати
	for _, e := range expenses {
		entry := &CashFlowEntry{
			ID:           e.ID,
			Date:         e.ExpenseDate,
			Type:         "expense",
			Counterparty: "", // Буде заповнено в enrichEntries
			Description:  e.Description,
			Category:     "", // Буде заповнено в enrichEntries
			CategoryCode: "", // Буде заповнено в enrichEntries
			Debit:        0,
			Credit:       e.Amount,
		}
		entries = append(entries, entry)
		totalCredit += e.Amount
	}

	// Сортуємо по даті
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})

	// Розраховуємо баланс для кожного запису
	balance := 0.0 // Початковий баланс завжди 0 згідно вимог
	for _, entry := range entries {
		balance += entry.Debit - entry.Credit
		entry.Balance = balance
	}

	// Заповнюємо додаткові дані
	if err := uc.enrichEntries(ctx, entries); err != nil {
		return nil, fmt.Errorf("failed to enrich entries: %w", err)
	}

	return &ListCashFlowOutput{
		Entries:      entries,
		Total:        int64(len(entries)),
		TotalDebit:   totalDebit,
		TotalCredit:  totalCredit,
		StartBalance: 0, // Завжди 0 згідно вимог
		EndBalance:   balance,
	}, nil
}

// enrichEntries заповнює додаткові дані (контрагенти, категорії) та розподіляє по колонках
func (uc *ListCashFlowUseCase) enrichEntries(ctx context.Context, entries []*CashFlowEntry) error {
	// Кешуємо дані для мінімізації запитів до БД
	ownershipCache := make(map[int64]string)
	categoryCache := make(map[int64]string)
	contractorCache := make(map[int64]string)
	categoryCodeCache := make(map[int64]string)

	for _, entry := range entries {
		if entry.Type == "payment" {
			// Отримуємо інформацію про платника (власника)
			payment, err := uc.paymentRepo.GetByID(ctx, entry.ID)
			if err == nil && payment != nil {
				entry.Counterparty = uc.getOwnerLastName(ctx, payment.OwnershipShareID, ownershipCache)
			}
		} else if entry.Type == "contractor_payment" {
			// Отримуємо інформацію про контрагента
			// TODO: Додати логіку отримання назви контрагента
			entry.Counterparty = "Контрагент"
		} else if entry.Type == "expense" {
			// Отримуємо інформацію про витрату
			expense, err := uc.expenseRepo.GetByID(ctx, entry.ID)
			if err == nil && expense != nil {
				// Отримуємо категорію
				entry.Category = uc.getCategoryInfo(ctx, expense.CategoryID, categoryCache)
				categoryCode := uc.getCategoryCode(ctx, expense.CategoryID, categoryCodeCache)
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
					entry.Counterparty = uc.getContractorInfo(ctx, expense.ContractorID, contractorCache)
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
func (uc *ListCashFlowUseCase) getOwnerLastName(ctx context.Context, ownershipID int64, cache map[int64]string) string {
	if name, ok := cache[ownershipID]; ok {
		return name
	}

	ownership, err := uc.ownershipRepo.GetByID(ctx, ownershipID)
	if err != nil || ownership == nil {
		cache[ownershipID] = "Невідомо"
		return "Невідомо"
	}

	owner, _ := uc.ownerRepo.GetByID(ctx, ownership.OwnerID)
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
func (uc *ListCashFlowUseCase) getCategoryInfo(ctx context.Context, categoryID int64, cache map[int64]string) string {
	if name, ok := cache[categoryID]; ok {
		return name
	}

	category, err := uc.categoryRepo.GetByID(ctx, categoryID)
	if err != nil || category == nil {
		cache[categoryID] = "Невідомо"
		return "Невідомо"
	}

	cache[categoryID] = category.Name
	return category.Name
}

// getCategoryCode отримує код категорії витрат
func (uc *ListCashFlowUseCase) getCategoryCode(ctx context.Context, categoryID int64, cache map[int64]string) string {
	if code, ok := cache[categoryID]; ok {
		return code
	}

	category, err := uc.categoryRepo.GetByID(ctx, categoryID)
	if err != nil || category == nil || category.Code == nil {
		cache[categoryID] = ""
		return ""
	}

	cache[categoryID] = *category.Code
	return *category.Code
}

// getContractorInfo отримує назву контрагента
func (uc *ListCashFlowUseCase) getContractorInfo(ctx context.Context, contractorID *int64, cache map[int64]string) string {
	if contractorID == nil {
		return "-"
	}

	if name, ok := cache[*contractorID]; ok {
		return name
	}

	contractor, err := uc.contractorRepo.GetByID(ctx, *contractorID)
	if err != nil || contractor == nil {
		cache[*contractorID] = "Невідомо"
		return "Невідомо"
	}

	cache[*contractorID] = contractor.Name
	return contractor.Name
}
