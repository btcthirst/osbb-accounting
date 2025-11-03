package domain

import (
	"fmt"
	"time"
)

// PaymentType представляє тип платежу.
type PaymentType string

const (
	// PaymentTypeIncoming - вхідний платіж (від мешканця)
	PaymentTypeIncoming PaymentType = "incoming"

	// PaymentTypeCharge - нарахування (борг)
	PaymentTypeCharge PaymentType = "charge"
)

// PaymentCategory представляє категорію платежу.
type PaymentCategory string

const (
	// CategoryMaintenance - утримання будинку
	CategoryMaintenance PaymentCategory = "maintenance"

	// CategoryRepairs - поточний ремонт
	CategoryRepairs PaymentCategory = "repairs"

	// CategoryUtilities - комунальні послуги
	CategoryUtilities PaymentCategory = "utilities"

	// CategoryHeating - опалення
	CategoryHeating PaymentCategory = "heating"

	// CategoryWater - водопостачання
	CategoryWater PaymentCategory = "water"

	// CategoryElectricity - електроенергія
	CategoryElectricity PaymentCategory = "electricity"

	// CategoryGas - газопостачання
	CategoryGas PaymentCategory = "gas"

	// CategoryOther - інше
	CategoryOther PaymentCategory = "other"
)

// Payment представляє платіж або нарахування для квартири.
type Payment struct {
	// ID - унікальний ідентифікатор платежу
	ID int `json:"id" db:"id"`

	// ApartmentID - ID квартири, до якої відноситься платіж
	ApartmentID int `json:"apartment_id" db:"apartment_id"`

	// Type - тип платежу (incoming/charge)
	Type PaymentType `json:"type" db:"type"`

	// Category - категорія платежу
	Category PaymentCategory `json:"category" db:"category"`

	// Amount - сума платежу (додатня для incoming, від'ємна для charge)
	Amount float64 `json:"amount" db:"amount"`

	// Description - опис платежу
	Description string `json:"description" db:"description"`

	// PaymentDate - дата здійснення платежу або нарахування
	PaymentDate time.Time `json:"payment_date" db:"payment_date"`

	// Period - період, за який нараховується (YYYY-MM, наприклад "2025-01")
	Period string `json:"period" db:"period"`

	// CreatedBy - ID користувача, який створив запис
	CreatedBy int `json:"created_by" db:"created_by"`

	// CreatedAt - час створення запису
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt - час останнього оновлення
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Notes - додаткові примітки
	Notes string `json:"notes,omitempty" db:"notes"`
}

// Validate виконує валідацію платежу.
func (p *Payment) Validate() error {
	if p.ApartmentID <= 0 {
		return ErrInvalidApartmentID
	}

	if p.Type != PaymentTypeIncoming && p.Type != PaymentTypeCharge {
		return ErrInvalidPaymentType
	}

	if p.Amount == 0 {
		return ErrInvalidAmount
	}

	// Для нарахувань сума має бути додатною (представляє борг)
	if p.Type == PaymentTypeCharge && p.Amount < 0 {
		return ErrInvalidAmount
	}

	// Для платежів сума має бути додатною (представляє оплату)
	if p.Type == PaymentTypeIncoming && p.Amount < 0 {
		return ErrInvalidAmount
	}

	if p.Description == "" {
		return ErrDescriptionRequired
	}

	if p.Period == "" {
		return ErrPeriodRequired
	}

	return nil
}

// IsIncoming перевіряє, чи є це вхідний платіж.
func (p *Payment) IsIncoming() bool {
	return p.Type == PaymentTypeIncoming
}

// IsCharge перевіряє, чи є це нарахування.
func (p *Payment) IsCharge() bool {
	return p.Type == PaymentTypeCharge
}

// GetSignedAmount повертає суму зі знаком для розрахунку балансу.
// Платежі додатні (+), нарахування від'ємні (-).
func (p *Payment) GetSignedAmount() float64 {
	if p.IsIncoming() {
		return p.Amount
	}
	return -p.Amount
}

// GetDisplayAmount повертає суму для відображення (завжди додатня).
func (p *Payment) GetDisplayAmount() float64 {
	if p.Amount < 0 {
		return -p.Amount
	}
	return p.Amount
}

// FormatPeriod форматує період для відображення.
func (p *Payment) FormatPeriod() string {
	// "2025-01" -> "Січень 2025"
	if len(p.Period) != 7 {
		return p.Period
	}

	monthNames := map[string]string{
		"01": "Січень", "02": "Лютий", "03": "Березень",
		"04": "Квітень", "05": "Травень", "06": "Червень",
		"07": "Липень", "08": "Серпень", "09": "Вересень",
		"10": "Жовтень", "11": "Листопад", "12": "Грудень",
	}

	year := p.Period[:4]
	month := p.Period[5:7]

	if monthName, ok := monthNames[month]; ok {
		return fmt.Sprintf("%s %s", monthName, year)
	}

	return p.Period
}

// GetCategoryDisplay повертає відображувану назву категорії.
func (p *Payment) GetCategoryDisplay() string {
	categories := map[PaymentCategory]string{
		CategoryMaintenance: "Утримання будинку",
		CategoryRepairs:     "Поточний ремонт",
		CategoryUtilities:   "Комунальні послуги",
		CategoryHeating:     "Опалення",
		CategoryWater:       "Водопостачання",
		CategoryElectricity: "Електроенергія",
		CategoryGas:         "Газопостачання",
		CategoryOther:       "Інше",
	}

	if display, ok := categories[p.Category]; ok {
		return display
	}
	return string(p.Category)
}

// PaymentSummary містить зведену інформацію про платежі квартири.
type PaymentSummary struct {
	ApartmentID   int        // ID квартири
	TotalCharges  float64    // Загальна сума нарахувань
	TotalPayments float64    // Загальна сума платежів
	Balance       float64    // Баланс (payments - charges)
	LastPayment   *time.Time // Дата останнього платежу
	PaymentsCount int        // Кількість платежів
	ChargesCount  int        // Кількість нарахувань
}

// HasDebt перевіряє, чи є заборгованість.
func (ps *PaymentSummary) HasDebt() bool {
	return ps.Balance < 0
}

// GetDebtAmount повертає суму боргу (завжди додатня).
func (ps *PaymentSummary) GetDebtAmount() float64 {
	if ps.Balance < 0 {
		return -ps.Balance
	}
	return 0
}

// GetOverpaymentAmount повертає суму переплати (завжди додатня).
func (ps *PaymentSummary) GetOverpaymentAmount() float64 {
	if ps.Balance > 0 {
		return ps.Balance
	}
	return 0
}

// PaymentFilter містить параметри для фільтрації платежів.
type PaymentFilter struct {
	// ApartmentID - фільтр по квартирі (nil = всі квартири)
	ApartmentID *int

	// Type - фільтр по типу (nil = всі типи)
	Type *PaymentType

	// Category - фільтр по категорії (nil = всі категорії)
	Category *PaymentCategory

	// Period - фільтр по періоду (nil = всі періоди)
	Period *string

	// DateFrom - початкова дата
	DateFrom *time.Time

	// DateTo - кінцева дата
	DateTo *time.Time

	// MinAmount - мінімальна сума
	MinAmount *float64

	// MaxAmount - максимальна сума
	MaxAmount *float64

	// CreatedBy - фільтр по користувачу, який створив
	CreatedBy *int

	// SortBy - поле для сортування
	SortBy string

	// SortDesc - сортування в зворотному порядку
	SortDesc bool
}

// DefaultPaymentFilter повертає фільтр за замовчуванням.
func DefaultPaymentFilter() *PaymentFilter {
	return &PaymentFilter{
		SortBy:   "payment_date",
		SortDesc: true, // Нові спочатку
	}
}

// MonthlyReport містить звіт за місяць для квартири.
type MonthlyReport struct {
	ApartmentID   int
	Period        string
	TotalCharges  float64
	TotalPayments float64
	Balance       float64
	PrevBalance   float64 // Баланс з попереднього періоду
	ByCategory    map[PaymentCategory]float64
}
