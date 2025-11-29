// domain/entity/contractor_payment.go
package entity

import (
	"errors"
	"time"
)

// ContractorPayment представляє платіж від контрагента (орендна плата, утримання мереж тощо).
type ContractorPayment struct {
	ID            int64
	ContractorID  int64
	PaymentDate   time.Time
	Amount        float64
	PaymentMethod PaymentMethod // Використовуємо той же enum, що й для Payment
	Purpose       string        // Призначення платежу

	// Період, за який здійснюється платіж (опціонально)
	PeriodMonth *int // 1-12
	PeriodYear  *int // Наприклад, 2024

	ReceiptNumber *string // Номер квитанції/документа
	Notes         *string

	// Soft Delete Pattern
	DeletedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validation errors
var (
	ErrContractorPaymentContractorRequired = errors.New("contractor ID is required")
	ErrContractorPaymentAmountInvalid      = errors.New("amount must be > 0")
	ErrContractorPaymentDateFuture         = errors.New("payment date cannot be in the future")
	ErrContractorPaymentMethodInvalid      = errors.New("invalid payment method")
	ErrContractorPaymentPeriodInvalid      = errors.New("invalid period (month must be 1-12, year must be realistic)")
	ErrContractorPaymentPurposeRequired    = errors.New("purpose is required")
)

// NewContractorPayment створює новий платіж від контрагента з валідацією.
func NewContractorPayment(
	contractorID int64,
	amount float64,
	paymentMethod PaymentMethod,
	purpose string,
	paymentDate time.Time,
) (*ContractorPayment, error) {
	payment := &ContractorPayment{
		ContractorID:  contractorID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Purpose:       purpose,
		PaymentDate:   paymentDate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := payment.Validate(); err != nil {
		return nil, err
	}

	return payment, nil
}

// Validate перевіряє коректність даних платежу.
func (cp *ContractorPayment) Validate() error {
	if cp.ContractorID <= 0 {
		return ErrContractorPaymentContractorRequired
	}

	if cp.Amount <= 0 {
		return ErrContractorPaymentAmountInvalid
	}

	if cp.PaymentDate.After(time.Now()) {
		return ErrContractorPaymentDateFuture
	}

	if !cp.PaymentMethod.IsValid() {
		return ErrContractorPaymentMethodInvalid
	}

	if cp.Purpose == "" {
		return ErrContractorPaymentPurposeRequired
	}

	// Перевірка періоду, якщо він вказаний
	if cp.PeriodMonth != nil {
		if *cp.PeriodMonth < 1 || *cp.PeriodMonth > 12 {
			return ErrContractorPaymentPeriodInvalid
		}
	}
	if cp.PeriodYear != nil {
		// Допускаємо діапазон від 2020 до поточний рік + 1
		currentYear := time.Now().Year()
		if *cp.PeriodYear < 2020 || *cp.PeriodYear > currentYear+1 {
			return ErrContractorPaymentPeriodInvalid
		}
	}

	return nil
}

// SetPeriod встановлює період оплати.
func (cp *ContractorPayment) SetPeriod(month, year int) error {
	cp.PeriodMonth = &month
	cp.PeriodYear = &year
	cp.UpdatedAt = time.Now()
	return cp.Validate()
}

// ClearPeriod очищує період оплати.
func (cp *ContractorPayment) ClearPeriod() {
	cp.PeriodMonth = nil
	cp.PeriodYear = nil
	cp.UpdatedAt = time.Now()
}

// Update оновлює основні дані платежу.
func (cp *ContractorPayment) Update(
	amount float64,
	method PaymentMethod,
	purpose string,
	date time.Time,
	receiptNumber *string,
	notes *string,
) error {
	cp.Amount = amount
	cp.PaymentMethod = method
	cp.Purpose = purpose
	cp.PaymentDate = date
	cp.ReceiptNumber = receiptNumber
	cp.Notes = notes
	cp.UpdatedAt = time.Now()

	return cp.Validate()
}

// SoftDelete позначає платіж як видалений.
func (cp *ContractorPayment) SoftDelete() {
	now := time.Now()
	cp.DeletedAt = &now
	cp.UpdatedAt = now
}

// IsDeleted перевіряє чи видалений запис.
func (cp *ContractorPayment) IsDeleted() bool {
	return cp.DeletedAt != nil
}
