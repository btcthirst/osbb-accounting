// domain/entity/charge.go
package entity

import (
	"errors"
	"time"
)

// Charge представляє нарахування по частці власності.
// Нарахування створюються щомісяця за різними категоріями (комунальні, утримання, тощо).
type Charge struct {
	ID               int64
	OwnershipShareID int64 // Зв'язок з ownership_shares
	ChargeType       ChargeType
	ChargeDate       time.Time
	PeriodMonth      int // Місяць нарахування (1-12)
	PeriodYear       int // Рік нарахування
	Amount           float64
	Tariff           *float64 // Тариф за одиницю (опціонально)
	Quantity         *float64 // Кількість одиниць (опціонально, напр. кВт/год)
	Description      *string
	Notes            *string
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ChargeType представляє тип нарахування.
type ChargeType string

const (
	ChargeTypeMaintenance ChargeType = "maintenance" // Утримання будинку
	ChargeTypeUtility     ChargeType = "utility"     // Комунальні послуги
	ChargeTypeRepair      ChargeType = "repair"      // Ремонт
	ChargeTypePenalty     ChargeType = "penalty"     // Пеня за прострочення
	ChargeTypeOther       ChargeType = "other"       // Інше
)

// Validation errors
var (
	ErrChargeOwnershipShareIDRequired = errors.New("ownership share ID is required")
	ErrChargeTypeInvalid              = errors.New("invalid charge type")
	ErrChargeDateRequired             = errors.New("charge date is required")
	ErrChargeDateInFuture             = errors.New("charge date cannot be in the future")
	ErrChargePeriodMonthInvalid       = errors.New("period month must be between 1 and 12")
	ErrChargePeriodYearInvalid        = errors.New("period year must be >= 2020 and <= 2100")
	ErrChargeAmountInvalid            = errors.New("amount must be >= 0")
	ErrChargeTariffInvalid            = errors.New("tariff must be > 0")
	ErrChargeQuantityInvalid          = errors.New("quantity must be > 0")
	ErrChargeCalculationInconsistent  = errors.New("if tariff and quantity provided, amount must equal tariff * quantity")
	ErrChargeDuplicatePeriod          = errors.New("charge for this period already exists")
)

// Константи валідації
const (
	MinChargeYear         = 2020
	MaxChargeYear         = 2100
	MaxChargeAmount       = 1000000.00 // Максимальна сума нарахування
	ChargeAmountPrecision = 0.01       // Точність для порівняння float
)

// NewCharge створює нове нарахування з валідацією.
func NewCharge(
	ownershipShareID int64,
	chargeType ChargeType,
	chargeDate time.Time,
	periodMonth, periodYear int,
	amount float64,
	tariff, quantity *float64,
	description *string,
) (*Charge, error) {
	charge := &Charge{
		OwnershipShareID: ownershipShareID,
		ChargeType:       chargeType,
		ChargeDate:       chargeDate,
		PeriodMonth:      periodMonth,
		PeriodYear:       periodYear,
		Amount:           amount,
		Tariff:           tariff,
		Quantity:         quantity,
		Description:      description,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := charge.Validate(); err != nil {
		return nil, err
	}

	return charge, nil
}

// Validate перевіряє коректність всіх полів нарахування.
func (c *Charge) Validate() error {
	// OwnershipShareID validation
	if c.OwnershipShareID <= 0 {
		return ErrChargeOwnershipShareIDRequired
	}
	// ChargeType validation
	if !c.ChargeType.IsValid() {
		return ErrChargeTypeInvalid
	}

	// ChargeDate validation
	if c.ChargeDate.IsZero() {
		return ErrChargeDateRequired
	}
	if c.ChargeDate.After(time.Now()) {
		return ErrChargeDateInFuture
	}

	// Period validation
	if c.PeriodMonth < 1 || c.PeriodMonth > 12 {
		return ErrChargePeriodMonthInvalid
	}
	if c.PeriodYear < MinChargeYear || c.PeriodYear > MaxChargeYear {
		return ErrChargePeriodYearInvalid
	}

	// Amount validation
	if c.Amount < 0 || c.Amount > MaxChargeAmount {
		return ErrChargeAmountInvalid
	}

	// Tariff validation (optional)
	if c.Tariff != nil && *c.Tariff <= 0 {
		return ErrChargeTariffInvalid
	}

	// Quantity validation (optional)
	if c.Quantity != nil && *c.Quantity <= 0 {
		return ErrChargeQuantityInvalid
	}

	// Перевірка консистентності: якщо вказані тариф та кількість,
	// сума має дорівнювати їх добутку (з точністю до копійок)
	if c.Tariff != nil && c.Quantity != nil {
		calculatedAmount := *c.Tariff * *c.Quantity
		if !floatEquals(c.Amount, calculatedAmount, ChargeAmountPrecision) {
			return ErrChargeCalculationInconsistent
		}
	}

	return nil
}

// IsValid перевіряє чи валідний тип нарахування.
func (ct ChargeType) IsValid() bool {
	switch ct {
	case ChargeTypeMaintenance, ChargeTypeUtility, ChargeTypeRepair,
		ChargeTypePenalty, ChargeTypeOther:
		return true
	default:
		return false
	}
}

// Update оновлює дані нарахування.
func (c *Charge) Update(
	amount float64,
	tariff, quantity *float64,
	description, notes *string,
) error {
	c.Amount = amount
	c.Tariff = tariff
	c.Quantity = quantity
	c.Description = description
	c.Notes = notes
	c.UpdatedAt = time.Now()
	return c.Validate()
}

// RecalculateAmount перераховує суму на основі тарифу та кількості.
func (c *Charge) RecalculateAmount() error {
	if c.Tariff == nil || c.Quantity == nil {
		return errors.New("tariff and quantity must be set to recalculate amount")
	}
	c.Amount = roundToTwoDecimals(*c.Tariff * *c.Quantity)
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// SoftDelete виконує м'яке видалення нарахування.
func (c *Charge) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}

// IsDeleted перевіряє чи видалене нарахування.
func (c *Charge) IsDeleted() bool {
	return c.DeletedAt != nil
}

// GetPeriodDisplay повертає відображення періоду (напр. "Січень 2024").
func (c *Charge) GetPeriodDisplay() string {
	months := []string{
		"", "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень",
		"Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	}
	if c.PeriodMonth >= 1 && c.PeriodMonth <= 12 {
		return months[c.PeriodMonth] + " " + string(rune(c.PeriodYear/1000+'0')) +
			string(rune((c.PeriodYear/100)%10+'0')) +
			string(rune((c.PeriodYear/10)%10+'0')) +
			string(rune(c.PeriodYear%10+'0'))
	}
	return "Невідомий період"
}

// GetPeriodShort повертає короткий формат періоду (напр. "01/2024").
func (c *Charge) GetPeriodShort() string {
	monthStr := string(rune(c.PeriodMonth/10+'0')) + string(rune(c.PeriodMonth%10+'0'))
	yearStr := string(rune(c.PeriodYear/1000+'0')) +
		string(rune((c.PeriodYear/100)%10+'0')) +
		string(rune((c.PeriodYear/10)%10+'0')) +
		string(rune(c.PeriodYear%10+'0'))
	return monthStr + "/" + yearStr
}

// GetAmountDisplay повертає відформатовану суму.
func (c *Charge) GetAmountDisplay() string {
	// Форматування з двома знаками після коми
	intPart := int64(c.Amount)
	fracPart := int((c.Amount - float64(intPart)) * 100)
	intStr := formatInt64(intPart)
	fracStr := string(rune(fracPart/10+'0')) + string(rune(fracPart%10+'0'))

	return intStr + "." + fracStr + " грн"
}

// GetCalculationDetails повертає деталі розрахунку.
func (c *Charge) GetCalculationDetails() string {
	if c.Tariff != nil && c.Quantity != nil {
		tariffStr := formatFloat64(*c.Tariff)
		quantityStr := formatFloat64(*c.Quantity)
		return quantityStr + " × " + tariffStr + " = " + c.GetAmountDisplay()
	}
	return c.GetAmountDisplay()
}

// GetTypeName повертає локалізовану назву типу нарахування.
func (ct ChargeType) GetDisplayName() string {
	switch ct {
	case ChargeTypeMaintenance:
		return "Утримання"
	case ChargeTypeUtility:
		return "Комунальні послуги"
	case ChargeTypeRepair:
		return "Ремонт"
	case ChargeTypePenalty:
		return "Пеня"
	case ChargeTypeOther:
		return "Інше"
	default:
		return "Невідомо"
	}
}

// IsPenalty перевіряє чи є нарахування пенею.
func (c *Charge) IsPenalty() bool {
	return c.ChargeType == ChargeTypePenalty
}

// IsForPeriod перевіряє чи належить нарахування до конкретного періоду.
func (c *Charge) IsForPeriod(month, year int) bool {
	return c.PeriodMonth == month && c.PeriodYear == year
}

// IsForCurrentPeriod перевіряє чи належить нарахування до поточного місяця.
func (c *Charge) IsForCurrentPeriod() bool {
	now := time.Now()
	return c.PeriodMonth == int(now.Month()) && c.PeriodYear == now.Year()
}

// GetPeriodStart повертає початок періоду нарахування.
func (c *Charge) GetPeriodStart() time.Time {
	return time.Date(c.PeriodYear, time.Month(c.PeriodMonth), 1, 0, 0, 0, 0, time.UTC)
}

// GetPeriodEnd повертає кінець періоду нарахування.
func (c *Charge) GetPeriodEnd() time.Time {
	start := c.GetPeriodStart()
	return start.AddDate(0, 1, -1) // Останній день місяця
}

// IsOverdue перевіряє чи прострочене нарахування (минув місяць від кінця періоду).
func (c *Charge) IsOverdue() bool {
	periodEnd := c.GetPeriodEnd()
	dueDate := periodEnd.AddDate(0, 1, 0) // Термін оплати - місяць після періоду
	return time.Now().After(dueDate)
}

// GetOverdueDays повертає кількість днів прострочення.
func (c *Charge) GetOverdueDays() int {
	if !c.IsOverdue() {
		return 0
	}
	periodEnd := c.GetPeriodEnd()
	dueDate := periodEnd.AddDate(0, 1, 0)
	overdueDuration := time.Since(dueDate)

	return int(overdueDuration.Hours() / 24)
}

// CalculatePenalty розраховує пеню за прострочення (0.1% від суми за кожен день).
func (c *Charge) CalculatePenalty(penaltyRate float64) float64 {
	if c.IsPenalty() {
		return 0 // Пеня не нараховується на пеню
	}
	overdueDays := c.GetOverdueDays()
	if overdueDays <= 0 {
		return 0
	}

	// penaltyRate - відсоток за день (напр. 0.001 для 0.1%)
	penalty := c.Amount * penaltyRate * float64(overdueDays)
	return roundToTwoDecimals(penalty)
}

// ============================================================================
// Helper функції
// ============================================================================
// floatEquals порівнює два float з точністю.
func floatEquals(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// roundToTwoDecimals округлює до двох знаків після коми.
func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

// formatInt64 форматує int64 з розділенням розрядів.
func formatInt64(n int64) string {
	if n < 0 {
		return "-" + formatInt64(-n)
	}
	if n < 1000 {
		// Перетворення int64 в string без strconv
		if n == 0 {
			return "0"
		}
		result := ""
		for n > 0 {
			result = string(rune(n%10+'0')) + result
			n /= 10
		}
		return result
	}
	// Рекурсивно форматуємо з пробілами
	return formatInt64(n/1000) + " " + formatThreeDigits(int(n%1000))
}

// formatThreeDigits форматує три цифри з ведучими нулями.
func formatThreeDigits(n int) string {
	if n < 10 {
		return "00" + string(rune(n+'0'))
	}
	if n < 100 {
		return "0" + string(rune(n/10+'0')) + string(rune(n%10+'0'))
	}
	return string(rune(n/100+'0')) + string(rune((n/10)%10+'0')) + string(rune(n%10+'0'))
}

// formatFloat64 форматує float64 з двома знаками після коми.
func formatFloat64(f float64) string {
	intPart := int64(f)
	fracPart := int((f - float64(intPart)) * 100)
	intStr := formatInt64(intPart)
	fracStr := string(rune(fracPart/10+'0')) + string(rune(fracPart%10+'0'))

	return intStr + "." + fracStr
}

// ChargePeriod представляє період нарахування (допоміжна структура).
type ChargePeriod struct {
	Month int
	Year  int
}

// NewChargePeriod створює новий період.
func NewChargePeriod(month, year int) (*ChargePeriod, error) {
	if month < 1 || month > 12 {
		return nil, ErrChargePeriodMonthInvalid
	}
	if year < MinChargeYear || year > MaxChargeYear {
		return nil, ErrChargePeriodYearInvalid
	}
	return &ChargePeriod{
		Month: month,
		Year:  year,
	}, nil
}

// CurrentPeriod повертає поточний період.
func CurrentPeriod() *ChargePeriod {
	now := time.Now()
	return &ChargePeriod{
		Month: int(now.Month()),
		Year:  now.Year(),
	}
}

// PreviousPeriod повертає попередній період.
func (cp *ChargePeriod) Previous() *ChargePeriod {
	month := cp.Month - 1
	year := cp.Year
	if month < 1 {
		month = 12
		year--
	}

	return &ChargePeriod{Month: month, Year: year}
}

// NextPeriod повертає наступний період.
func (cp *ChargePeriod) Next() *ChargePeriod {
	month := cp.Month + 1
	year := cp.Year
	if month > 12 {
		month = 1
		year++
	}

	return &ChargePeriod{Month: month, Year: year}
}

// Equals порівнює два періоди.
func (cp *ChargePeriod) Equals(other *ChargePeriod) bool {
	return cp.Month == other.Month && cp.Year == other.Year
}

// String повертає string representation періоду.
func (cp *ChargePeriod) String() string {
	monthStr := string(rune(cp.Month/10+'0')) + string(rune(cp.Month%10+'0'))
	yearStr := string(rune(cp.Year/1000+'0')) +
		string(rune((cp.Year/100)%10+'0')) +
		string(rune((cp.Year/10)%10+'0')) +
		string(rune(cp.Year%10+'0'))
	return monthStr + "/" + yearStr
}
