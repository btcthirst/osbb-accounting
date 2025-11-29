// application/usecase/cashflow/cashflow_usecase.go
package cashflow

import (
	"context"
	"sort"
	"time"

	"osbb-accounting/domain/repository"
)

// CashFlowEntry представляє єдину операцію в русі коштів
type CashFlowEntry struct {
	ID           int64
	Date         time.Time
	Type         string // "payment", "contractor_payment" або "expense"
	Counterparty string // Платник або Постачальник
	Description  string
	Category     string  // Категорія (для витрат)
	CategoryCode string  // Код категорії (313, 63, 641, тощо)
	Debit        float64 // Надходження
	Credit       float64 // Витрати
	Balance      float64 // Поточний баланс (розраховується)

	// Розбивка витрат по категоріях (для колонок 6-11)
	Credit313  float64 // Кошти на картку
	Credit63   float64 // Електроенергія
	Credit641  float64 // ПДФО 18%
	Credit6411 float64 // Військовий збір
	Credit651  float64 // ЄСВ 22%
	Credit94   float64 // Комісія банку
}

// ListCashFlowInput - вхідні дані для отримання списку операцій
type ListCashFlowInput struct {
	CurrentUserID int64
	StartDate     *time.Time
	EndDate       *time.Time
	Month         *int // 1-12
	Year          *int
	Limit         int
	Offset        int
}

// ListCashFlowOutput - результат отримання списку операцій
type ListCashFlowOutput struct {
	Entries      []*CashFlowEntry
	Total        int64
	TotalDebit   float64 // Загальна сума надходжень
	TotalCredit  float64 // Загальна сума витрат
	StartBalance float64 // Баланс на початок періоду (завжди 0 згідно вимог)
	EndBalance   float64 // Баланс на кінець періоду
}

// CashFlowUseCase - use case для роботи з рухом коштів
type CashFlowUseCase struct {
	paymentRepo           repository.PaymentRepository
	contractorPaymentRepo repository.ContractorPaymentRepository
	expenseRepo           repository.ExpenseRepository
	ownershipRepo         repository.OwnershipShareRepository
	categoryRepo          repository.ExpenseCategoryRepository
	contractorRepo        repository.ContractorRepository
}

// NewCashFlowUseCase створює новий use case
func NewCashFlowUseCase(
	paymentRepo repository.PaymentRepository,
	contractorPaymentRepo repository.ContractorPaymentRepository,
	expenseRepo repository.ExpenseRepository,
	ownershipRepo repository.OwnershipShareRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	contractorRepo repository.ContractorRepository,
) *CashFlowUseCase {
	return &CashFlowUseCase{
		paymentRepo:           paymentRepo,
		contractorPaymentRepo: contractorPaymentRepo,
		expenseRepo:           expenseRepo,
		ownershipRepo:         ownershipRepo,
		categoryRepo:          categoryRepo,
		contractorRepo:        contractorRepo,
	}
}

// List отримує список операцій руху коштів
func (uc *CashFlowUseCase) List(ctx context.Context, input ListCashFlowInput) (*ListCashFlowOutput, error) {
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
			Counterparty: "", // Буде заповнено в сервісі
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
			Counterparty: "", // Буде заповнено в сервісі
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
			Counterparty: "", // Буде заповнено в сервісі
			Description:  e.Description,
			Category:     "", // Буде заповнено в сервісі
			CategoryCode: "", // Буде заповнено в сервісі
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

	return &ListCashFlowOutput{
		Entries:      entries,
		Total:        int64(len(entries)),
		TotalDebit:   totalDebit,
		TotalCredit:  totalCredit,
		StartBalance: 0, // Завжди 0 згідно вимог
		EndBalance:   balance,
	}, nil
}
