// domain/entity/expense.go
package entity

import (
	"errors"
	"time"
)

// Expense представляє витрату ОСББ.
type Expense struct {
	ID           int64
	CategoryID   int64
	ContractorID *int64 // Опціонально - може бути витрата без контрагента
	ExpenseDate  time.Time
	Amount       float64
	Description  string

	// Документальне підтвердження
	DocumentType   *string // invoice, act, receipt, order, other
	DocumentNumber *string
	DocumentDate   *time.Time

	// Статус оплати
	PaymentStatus PaymentStatus
	PaidAmount    float64
	PaymentDate   *time.Time

	Notes *string

	// Затвердження (бізнес-логіка)
	ApprovedBy *int64
	ApprovedAt *time.Time

	// Soft Delete Pattern
	DeletedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PaymentStatus представляє статус оплати витрати.
type PaymentStatus string

const (
	PaymentStatusPending       PaymentStatus = "pending"        // Очікує оплати
	PaymentStatusPaid          PaymentStatus = "paid"           // Оплачено повністю
	PaymentStatusPartiallyPaid PaymentStatus = "partially_paid" // Оплачено частково
	PaymentStatusCancelled     PaymentStatus = "cancelled"      // Скасовано
)

// DocumentType представляє тип документа.
type DocumentType string

const (
	DocumentTypeInvoice DocumentType = "invoice" // Рахунок
	DocumentTypeAct     DocumentType = "act"     // Акт
	DocumentTypeReceipt DocumentType = "receipt" // Квитанція
	DocumentTypeOrder   DocumentType = "order"   // Наказ
	DocumentTypeOther   DocumentType = "other"   // Інше
)

// Validation errors
var (
	ErrExpenseCategoryIDRequired   = errors.New("category ID is required")
	ErrExpenseAmountInvalid        = errors.New("amount must be > 0")
	ErrExpenseDescriptionRequired  = errors.New("description is required")
	ErrExpenseDescriptionTooShort  = errors.New("description must be at least 3 characters")
	ErrExpenseDateInvalid          = errors.New("expense date cannot be in the future")
	ErrExpensePaymentStatusInvalid = errors.New("invalid payment status")
	ErrExpensePaidAmountInvalid    = errors.New("paid amount must be >= 0 and <= total amount")
	ErrExpenseDocumentDateInvalid  = errors.New("document date cannot be in the future")
	ErrExpensePaymentDateInvalid   = errors.New("payment date cannot be in the future")
	ErrExpenseAlreadyApproved      = errors.New("expense is already approved")
	ErrExpenseNotApproved          = errors.New("expense is not approved")
)

// NewExpense створює нову витрату з валідацією.
func NewExpense(
	categoryID int64,
	contractorID *int64,
	expenseDate time.Time,
	amount float64,
	description string,
) (*Expense, error) {
	expense := &Expense{
		CategoryID:    categoryID,
		ContractorID:  contractorID,
		ExpenseDate:   expenseDate,
		Amount:        amount,
		Description:   description,
		PaymentStatus: PaymentStatusPending,
		PaidAmount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := expense.Validate(); err != nil {
		return nil, err
	}

	return expense, nil
}

// Validate перевіряє коректність всіх полів витрати.
func (e *Expense) Validate() error {
	// Category ID validation
	if e.CategoryID <= 0 {
		return ErrExpenseCategoryIDRequired
	}

	// Amount validation
	if e.Amount <= 0 {
		return ErrExpenseAmountInvalid
	}

	// Description validation
	if e.Description == "" {
		return ErrExpenseDescriptionRequired
	}
	if len(e.Description) < 3 {
		return ErrExpenseDescriptionTooShort
	}

	// Expense date validation (не може бути у майбутньому)
	if e.ExpenseDate.After(time.Now()) {
		return ErrExpenseDateInvalid
	}

	// Payment status validation
	if !e.PaymentStatus.IsValid() {
		return ErrExpensePaymentStatusInvalid
	}

	// Paid amount validation
	if e.PaidAmount < 0 || e.PaidAmount > e.Amount {
		return ErrExpensePaidAmountInvalid
	}

	// Document date validation (опціонально)
	if e.DocumentDate != nil && e.DocumentDate.After(time.Now()) {
		return ErrExpenseDocumentDateInvalid
	}

	// Payment date validation (опціонально)
	if e.PaymentDate != nil && e.PaymentDate.After(time.Now()) {
		return ErrExpensePaymentDateInvalid
	}

	return nil
}

// IsValid перевіряє чи валідний статус оплати.
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusPaid, PaymentStatusPartiallyPaid, PaymentStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValid перевіряє чи валідний тип документа.
func (dt DocumentType) IsValid() bool {
	switch dt {
	case DocumentTypeInvoice, DocumentTypeAct, DocumentTypeReceipt, DocumentTypeOrder, DocumentTypeOther:
		return true
	default:
		return false
	}
}

// Update оновлює дані витрати.
func (e *Expense) Update(
	categoryID int64,
	contractorID *int64,
	expenseDate time.Time,
	amount float64,
	description string,
	notes *string,
) error {
	e.CategoryID = categoryID
	e.ContractorID = contractorID
	e.ExpenseDate = expenseDate
	e.Amount = amount
	e.Description = description
	e.Notes = notes
	e.UpdatedAt = time.Now()

	return e.Validate()
}

// SetDocument встановлює документальне підтвердження.
func (e *Expense) SetDocument(
	documentType DocumentType,
	documentNumber string,
	documentDate time.Time,
) error {
	if !documentType.IsValid() {
		return errors.New("invalid document type")
	}

	dt := string(documentType)
	e.DocumentType = &dt
	e.DocumentNumber = &documentNumber
	e.DocumentDate = &documentDate
	e.UpdatedAt = time.Now()

	return e.Validate()
}

// RecordPayment реєструє оплату витрати.
func (e *Expense) RecordPayment(amount float64, paymentDate time.Time) error {
	if amount <= 0 {
		return errors.New("payment amount must be > 0")
	}

	newPaidAmount := e.PaidAmount + amount
	if newPaidAmount > e.Amount {
		return errors.New("total paid amount would exceed expense amount")
	}

	e.PaidAmount = newPaidAmount
	e.PaymentDate = &paymentDate
	e.UpdatedAt = time.Now()

	// Автоматично оновлюємо статус
	if e.PaidAmount >= e.Amount {
		e.PaymentStatus = PaymentStatusPaid
	} else if e.PaidAmount > 0 {
		e.PaymentStatus = PaymentStatusPartiallyPaid
	}

	return nil
}

// Approve затверджує витрату.
func (e *Expense) Approve(approverID int64) error {
	if e.IsApproved() {
		return ErrExpenseAlreadyApproved
	}

	now := time.Now()
	e.ApprovedBy = &approverID
	e.ApprovedAt = &now
	e.UpdatedAt = now

	return nil
}

// Unapprove скасовує затвердження витрати.
func (e *Expense) Unapprove() error {
	if !e.IsApproved() {
		return ErrExpenseNotApproved
	}

	e.ApprovedBy = nil
	e.ApprovedAt = nil
	e.UpdatedAt = time.Now()

	return nil
}

// Cancel скасовує витрату.
func (e *Expense) Cancel() {
	e.PaymentStatus = PaymentStatusCancelled
	e.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення витрати.
func (e *Expense) SoftDelete() {
	now := time.Now()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

// IsDeleted перевіряє чи видалена витрата.
func (e *Expense) IsDeleted() bool {
	return e.DeletedAt != nil
}

// IsApproved перевіряє чи затверджена витрата.
func (e *Expense) IsApproved() bool {
	return e.ApprovedBy != nil && e.ApprovedAt != nil
}

// IsPaid перевіряє чи оплачена витрата повністю.
func (e *Expense) IsPaid() bool {
	return e.PaymentStatus == PaymentStatusPaid
}

// IsPartiallyPaid перевіряє чи оплачена витрата частково.
func (e *Expense) IsPartiallyPaid() bool {
	return e.PaymentStatus == PaymentStatusPartiallyPaid
}

// IsCancelled перевіряє чи скасована витрата.
func (e *Expense) IsCancelled() bool {
	return e.PaymentStatus == PaymentStatusCancelled
}

// GetRemainingAmount повертає залишок до оплати.
func (e *Expense) GetRemainingAmount() float64 {
	return e.Amount - e.PaidAmount
}

// GetPaymentProgress повертає прогрес оплати у відсотках.
func (e *Expense) GetPaymentProgress() float64 {
	if e.Amount == 0 {
		return 0
	}
	return (e.PaidAmount / e.Amount) * 100
}

// GetStatusName повертає локалізовану назву статусу оплати.
func (ps PaymentStatus) GetDisplayName() string {
	switch ps {
	case PaymentStatusPending:
		return "Очікує оплати"
	case PaymentStatusPaid:
		return "Оплачено"
	case PaymentStatusPartiallyPaid:
		return "Оплачено частково"
	case PaymentStatusCancelled:
		return "Скасовано"
	default:
		return "Невідомо"
	}
}

// GetDocumentTypeName повертає локалізовану назву типу документа.
func (dt DocumentType) GetDisplayName() string {
	switch dt {
	case DocumentTypeInvoice:
		return "Рахунок"
	case DocumentTypeAct:
		return "Акт"
	case DocumentTypeReceipt:
		return "Квитанція"
	case DocumentTypeOrder:
		return "Наказ"
	case DocumentTypeOther:
		return "Інше"
	default:
		return "Невідомо"
	}
}
