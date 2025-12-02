package cashflow

import "time"

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
