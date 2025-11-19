// domain/entity/payment.go
package entity

import (
	"errors"
	"time"
)

// Payment представляє платіж від власника (внесок, оплата послуг тощо).
type Payment struct {
	ID               int64
	OwnershipShareID int64
	PaymentDate      time.Time
	Amount           float64
	PaymentMethod    PaymentMethod
	PaymentPurpose   string // Призначення платежу

	// Період, за який здійснюється платіж (опціонально)
	PeriodMonth *int // 1-12
	PeriodYear  *int // Наприклад, 2024

	ReceiptNumber *string // Номер квитанції/документа
	Notes         *string

	// Затвердження (контроль бухгалтера/голови)
	ApprovedBy *int64
	ApprovedAt *time.Time

	// Soft Delete Pattern
	DeletedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PaymentMethod визначає спосіб оплати.
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"          // Готівка
	PaymentMethodCard         PaymentMethod = "card"          // Картка/Термінал
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer" // Банківський переказ (IBAN)
	PaymentMethodOther        PaymentMethod = "other"         // Інше
)

// Validation errors
var (
	ErrPaymentOwnershipRequired = errors.New("ownership share ID is required")
	ErrPaymentAmountInvalid     = errors.New("amount must be > 0")
	ErrPaymentDateFuture        = errors.New("payment date cannot be in the future")
	ErrPaymentMethodInvalid     = errors.New("invalid payment method")
	ErrPaymentPeriodInvalid     = errors.New("invalid period (month must be 1-12, year must be realistic)")
	ErrPaymentAlreadyApproved   = errors.New("payment is already approved")
	ErrPaymentNotApproved       = errors.New("payment is not approved")
)

// NewPayment створює новий платіж з валідацією.
func NewPayment(
	ownershipShareID int64,
	amount float64,
	paymentMethod PaymentMethod,
	paymentPurpose string,
	paymentDate time.Time,
) (*Payment, error) {
	payment := &Payment{
		OwnershipShareID: ownershipShareID,
		Amount:           amount,
		PaymentMethod:    paymentMethod,
		PaymentPurpose:   paymentPurpose,
		PaymentDate:      paymentDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := payment.Validate(); err != nil {
		return nil, err
	}

	return payment, nil
}

// Validate перевіряє коректність даних платежу.
func (p *Payment) Validate() error {
	if p.OwnershipShareID <= 0 {
		return ErrPaymentOwnershipRequired
	}

	if p.Amount <= 0 {
		return ErrPaymentAmountInvalid
	}

	if p.PaymentDate.After(time.Now()) {
		return ErrPaymentDateFuture
	}

	if !p.PaymentMethod.IsValid() {
		return ErrPaymentMethodInvalid
	}

	// Перевірка періоду, якщо він вказаний
	if p.PeriodMonth != nil {
		if *p.PeriodMonth < 1 || *p.PeriodMonth > 12 {
			return ErrPaymentPeriodInvalid
		}
	}
	if p.PeriodYear != nil {
		// Допускаємо діапазон від 2020 до поточний рік + 1
		currentYear := time.Now().Year()
		if *p.PeriodYear < 2020 || *p.PeriodYear > currentYear+1 {
			return ErrPaymentPeriodInvalid
		}
	}

	return nil
}

// IsValid перевіряє чи валідний метод оплати.
func (pm PaymentMethod) IsValid() bool {
	switch pm {
	case PaymentMethodCash, PaymentMethodCard, PaymentMethodBankTransfer, PaymentMethodOther:
		return true
	default:
		return false
	}
}

// SetPeriod встановлює період оплати.
func (p *Payment) SetPeriod(month, year int) error {
	p.PeriodMonth = &month
	p.PeriodYear = &year
	p.UpdatedAt = time.Now()
	return p.Validate()
}

// ClearPeriod очищує період оплати (якщо платіж не прив'язаний до періоду).
func (p *Payment) ClearPeriod() {
	p.PeriodMonth = nil
	p.PeriodYear = nil
	p.UpdatedAt = time.Now()
}

// Update оновлює основні дані платежу.
func (p *Payment) Update(
	amount float64,
	method PaymentMethod,
	purpose string,
	date time.Time,
	receiptNumber *string,
	notes *string,
) error {
	p.Amount = amount
	p.PaymentMethod = method
	p.PaymentPurpose = purpose
	p.PaymentDate = date
	p.ReceiptNumber = receiptNumber
	p.Notes = notes
	p.UpdatedAt = time.Now()

	return p.Validate()
}

// Approve підтверджує платіж (фіксація в бухгалтерії).
func (p *Payment) Approve(approverID int64) error {
	if p.IsApproved() {
		return ErrPaymentAlreadyApproved
	}

	now := time.Now()
	p.ApprovedBy = &approverID
	p.ApprovedAt = &now
	p.UpdatedAt = now

	return nil
}

// Unapprove знімає підтвердження платежу (для корегування).
func (p *Payment) Unapprove() error {
	if !p.IsApproved() {
		return ErrPaymentNotApproved
	}

	p.ApprovedBy = nil
	p.ApprovedAt = nil
	p.UpdatedAt = time.Now()

	return nil
}

// IsApproved перевіряє статус затвердження.
func (p *Payment) IsApproved() bool {
	return p.ApprovedBy != nil && p.ApprovedAt != nil
}

// SoftDelete позначає платіж як видалений.
func (p *Payment) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// IsDeleted перевіряє чи видалений запис.
func (p *Payment) IsDeleted() bool {
	return p.DeletedAt != nil
}

// GetMethodDisplayName повертає читабельну назву методу оплати.
func (pm PaymentMethod) GetDisplayName() string {
	switch pm {
	case PaymentMethodCash:
		return "Готівка"
	case PaymentMethodCard:
		return "Картка/Термінал"
	case PaymentMethodBankTransfer:
		return "Банківський переказ"
	case PaymentMethodOther:
		return "Інше"
	default:
		return "Невідомо"
	}
}
