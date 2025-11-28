// domain/entity/imported_monthly_record.go
package entity

import (
	"errors"
	"fmt"
	"time"
)

// ImportedMonthlyRecord представляє імпортований місячний запис з XLSX.
// Зберігає всі дані з файлу для можливості перегляду та міграції.
type ImportedMonthlyRecord struct {
	ID            int64
	ImportBatchID int64

	// Період
	PeriodMonth int // 1-12
	PeriodYear  int
	SheetName   *string // Назва аркушу (січень, лютий тощо)

	// Дані з XLSX
	ApartmentNumber string
	OwnerName       string
	AccountNumber   *string

	// Баланси
	OpeningDebit  float64
	OpeningCredit float64
	ClosingDebit  float64
	ClosingCredit float64

	// Інформація про квартиру
	TotalArea       *float64
	DiscountArea    float64
	DiscountPercent int

	// Нарахування
	Tariff         *float64
	ChargeAmount   float64
	DiscountAmount float64
	AmountDue      float64

	// Платежі
	AmountPaid float64

	// Інше
	Corrections *float64
	Notes       *string

	// Зв'язки з операційними таблицями
	ApartmentID      *int64
	OwnerID          *int64
	OwnershipShareID *int64
	ChargeID         *int64
	PaymentID        *int64

	// Статус міграції
	IsMigrated bool
	MigratedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validation errors
var (
	ErrImportedRecordBatchRequired     = errors.New("import batch ID is required")
	ErrImportedRecordPeriodInvalid     = errors.New("invalid period (month 1-12, year >= 2020)")
	ErrImportedRecordApartmentRequired = errors.New("apartment number is required")
	ErrImportedRecordOwnerRequired     = errors.New("owner name is required")
	ErrImportedRecordNegativeAmount    = errors.New("amounts cannot be negative")
	ErrImportedRecordDiscountInvalid   = errors.New("discount percent must be 0-100")
)

// NewImportedMonthlyRecord створює новий імпортований запис з валідацією.
func NewImportedMonthlyRecord(
	importBatchID int64,
	periodMonth, periodYear int,
	apartmentNumber, ownerName string,
) (*ImportedMonthlyRecord, error) {
	record := &ImportedMonthlyRecord{
		ImportBatchID:   importBatchID,
		PeriodMonth:     periodMonth,
		PeriodYear:      periodYear,
		ApartmentNumber: apartmentNumber,
		OwnerName:       ownerName,
		IsMigrated:      false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := record.Validate(); err != nil {
		return nil, err
	}

	return record, nil
}

// Validate перевіряє коректність даних запису.
func (r *ImportedMonthlyRecord) Validate() error {
	if r.ImportBatchID <= 0 {
		return ErrImportedRecordBatchRequired
	}

	if r.PeriodMonth < 1 || r.PeriodMonth > 12 {
		return ErrImportedRecordPeriodInvalid
	}

	if r.PeriodYear < 2020 || r.PeriodYear > 2100 {
		return ErrImportedRecordPeriodInvalid
	}

	if r.ApartmentNumber == "" {
		return ErrImportedRecordApartmentRequired
	}

	if r.OwnerName == "" {
		return ErrImportedRecordOwnerRequired
	}

	// Перевірка що суми не від'ємні
	if r.OpeningDebit < 0 || r.OpeningCredit < 0 ||
		r.ClosingDebit < 0 || r.ClosingCredit < 0 ||
		r.ChargeAmount < 0 || r.DiscountAmount < 0 ||
		r.AmountDue < 0 || r.AmountPaid < 0 || r.DiscountArea < 0 {
		return ErrImportedRecordNegativeAmount
	}

	if r.DiscountPercent < 0 || r.DiscountPercent > 100 {
		return ErrImportedRecordDiscountInvalid
	}

	return nil
}

// GetPeriodDisplay повертає відображення періоду (напр. "Січень 2023").
func (r *ImportedMonthlyRecord) GetPeriodDisplay() string {
	months := []string{
		"", "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень",
		"Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	}

	if r.PeriodMonth >= 1 && r.PeriodMonth <= 12 {
		return fmt.Sprintf("%s %d", months[r.PeriodMonth], r.PeriodYear)
	}

	return fmt.Sprintf("%02d/%d", r.PeriodMonth, r.PeriodYear)
}

// GetBalance повертає баланс (дебет - кредит).
// Позитивне значення = борг, негативне = переплата.
func (r *ImportedMonthlyRecord) GetOpeningBalance() float64 {
	return r.OpeningDebit - r.OpeningCredit
}

// GetClosingBalance повертає кінцевий баланс.
func (r *ImportedMonthlyRecord) GetClosingBalance() float64 {
	return r.ClosingDebit - r.ClosingCredit
}

// IsDebt перевіряє чи є борг на кінець періоду.
func (r *ImportedMonthlyRecord) IsDebt() bool {
	return r.ClosingDebit > r.ClosingCredit
}

// MarkAsMigrated позначає запис як перенесений в операційні таблиці.
func (r *ImportedMonthlyRecord) MarkAsMigrated(
	apartmentID, ownerID, ownershipShareID *int64,
	chargeID, paymentID *int64,
) {
	r.IsMigrated = true
	now := time.Now()
	r.MigratedAt = &now
	r.ApartmentID = apartmentID
	r.OwnerID = ownerID
	r.OwnershipShareID = ownershipShareID
	r.ChargeID = chargeID
	r.PaymentID = paymentID
	r.UpdatedAt = now
}

// UnmarkMigration скасовує міграцію (якщо потрібно відкотити).
func (r *ImportedMonthlyRecord) UnmarkMigration() {
	r.IsMigrated = false
	r.MigratedAt = nil
	r.ApartmentID = nil
	r.OwnerID = nil
	r.OwnershipShareID = nil
	r.ChargeID = nil
	r.PaymentID = nil
	r.UpdatedAt = time.Now()
}
