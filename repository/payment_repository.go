package repository

import (
	"osbb-accounting/domain"
	"time"
)

// PaymentRepository визначає контракт для всіх операцій з платежами.
// Дотримується Repository Pattern для ізоляції доступу до даних.
type PaymentRepository interface {
	// Save зберігає новий платіж в БД.
	//
	// Параметри:
	//   - payment: вказівник на domain.Payment для збереження
	//
	// Повертає:
	//   - error: nil при успіху
	Save(payment *domain.Payment) error

	// FindByID знаходить платіж за ID.
	//
	// Параметри:
	//   - id: унікальний ідентифікатор платежу
	//
	// Повертає:
	//   - *domain.Payment: знайдений платіж
	//   - error: nil при успіху, domain.ErrPaymentNotFound якщо не знайдено
	FindByID(id int) (*domain.Payment, error)

	// FindByApartmentID знаходить всі платежі квартири.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//
	// Повертає:
	//   - []*domain.Payment: список платежів
	//   - error: nil при успіху
	FindByApartmentID(apartmentID int) ([]*domain.Payment, error)

	// FindByFilter знаходить платежі за вказаними критеріями.
	//
	// Параметри:
	//   - filter: параметри фільтрації
	//
	// Повертає:
	//   - []*domain.Payment: знайдені платежі
	//   - error: nil при успіху
	FindByFilter(filter *domain.PaymentFilter) ([]*domain.Payment, error)

	// FindByPeriod знаходить всі платежі за період.
	//
	// Параметри:
	//   - period: період у форматі "YYYY-MM"
	//
	// Повертає:
	//   - []*domain.Payment: платежі за період
	//   - error: nil при успіху
	FindByPeriod(period string) ([]*domain.Payment, error)

	// FindByApartmentAndPeriod знаходить платежі квартири за період.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//   - period: період у форматі "YYYY-MM"
	//
	// Повертає:
	//   - []*domain.Payment: платежі за період
	//   - error: nil при успіху
	FindByApartmentAndPeriod(apartmentID int, period string) ([]*domain.Payment, error)

	// Update оновлює існуючий платіж.
	//
	// Параметри:
	//   - payment: вказівник на domain.Payment з оновленими даними
	//
	// Повертає:
	//   - error: nil при успіху
	Update(payment *domain.Payment) error

	// Delete видаляє платіж з БД.
	//
	// Параметри:
	//   - id: ідентифікатор платежу
	//
	// Повертає:
	//   - error: nil при успіху
	Delete(id int) error

	// GetSummaryByApartment повертає зведену інформацію про платежі квартири.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//
	// Повертає:
	//   - *domain.PaymentSummary: зведення
	//   - error: nil при успіху
	GetSummaryByApartment(apartmentID int) (*domain.PaymentSummary, error)

	// GetSummaryByPeriod повертає зведення за період для всіх квартир.
	//
	// Параметри:
	//   - period: період у форматі "YYYY-MM"
	//
	// Повертає:
	//   - map[int]*domain.PaymentSummary: зведення по квартирах (ключ - apartment_id)
	//   - error: nil при успіху
	GetSummaryByPeriod(period string) (map[int]*domain.PaymentSummary, error)

	// GetMonthlyReport генерує місячний звіт для квартири.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//   - period: період у форматі "YYYY-MM"
	//
	// Повертає:
	//   - *domain.MonthlyReport: звіт за місяць
	//   - error: nil при успіху
	GetMonthlyReport(apartmentID int, period string) (*domain.MonthlyReport, error)

	// GetDebtorApartments повертає список квартир з боргом.
	//
	// Повертає:
	//   - []int: ID квартир з боргом
	//   - error: nil при успіху
	GetDebtorApartments() ([]int, error)

	// CalculateBalance розраховує поточний баланс квартири.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//
	// Повертає:
	//   - float64: баланс (додатній = переплата, від'ємний = борг)
	//   - error: nil при успіху
	CalculateBalance(apartmentID int) (float64, error)

	// CalculateBalanceForPeriod розраховує баланс квартири за період.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//   - period: період у форматі "YYYY-MM"
	//
	// Повертає:
	//   - float64: баланс за період
	//   - error: nil при успіху
	CalculateBalanceForPeriod(apartmentID int, period string) (float64, error)

	// GetTotalCharges повертає загальну суму нарахувань.
	//
	// Параметри:
	//   - apartmentID: ID квартири (опціонально, 0 = всі квартири)
	//
	// Повертає:
	//   - float64: загальна сума нарахувань
	//   - error: nil при успіху
	GetTotalCharges(apartmentID int) (float64, error)

	// GetTotalPayments повертає загальну суму платежів.
	//
	// Параметри:
	//   - apartmentID: ID квартири (опціонально, 0 = всі квартири)
	//
	// Повертає:
	//   - float64: загальна сума платежів
	//   - error: nil при успіху
	GetTotalPayments(apartmentID int) (float64, error)

	// GetPaymentsByCategory повертає платежі згруповані за категоріями.
	//
	// Параметри:
	//   - apartmentID: ID квартири (опціонально, 0 = всі квартири)
	//   - period: період (опціонально, "" = всі періоди)
	//
	// Повертає:
	//   - map[domain.PaymentCategory]float64: суми по категоріях
	//   - error: nil при успіху
	GetPaymentsByCategory(apartmentID int, period string) (map[domain.PaymentCategory]float64, error)

	// GetRecentPayments повертає останні платежі.
	//
	// Параметри:
	//   - limit: кількість записів
	//
	// Повертає:
	//   - []*domain.Payment: останні платежі
	//   - error: nil при успіху
	GetRecentPayments(limit int) ([]*domain.Payment, error)

	// Count підраховує кількість платежів.
	//
	// Параметри:
	//   - apartmentID: ID квартири (опціонально, 0 = всі квартири)
	//
	// Повертає:
	//   - int: кількість платежів
	//   - error: nil при успіху
	Count(apartmentID int) (int, error)
}

// PaymentStatistics містить загальну статистику по платежах.
type PaymentStatistics struct {
	TotalPayments      float64    // Загальна сума платежів
	TotalCharges       float64    // Загальна сума нарахувань
	TotalBalance       float64    // Загальний баланс
	TotalDebt          float64    // Загальна заборгованість
	ApartmentsWithDebt int        // Кількість квартир з боргом
	TotalPaymentsCount int        // Кількість платежів
	TotalChargesCount  int        // Кількість нарахувань
	LastPaymentDate    *time.Time // Дата останнього платежу
}
