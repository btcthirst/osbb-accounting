package payment

import (
	"time"

	"osbb-accounting/domain/entity"
)

// PaymentOutput - результат операцій з платежем.
type PaymentOutput struct {
	ID               int64
	OwnershipShareID int64
	PaymentDate      time.Time
	Amount           float64
	PaymentMethod    string
	PaymentPurpose   string
	PeriodMonth      *int
	PeriodYear       *int
	ReceiptNumber    *string
	Notes            *string
	ApprovedBy       *int64
	ApprovedAt       *time.Time
	IsApproved       bool
	MethodName       string
}

// mapPaymentToOutput - helper function для конвертації entity в output.
func mapPaymentToOutput(p *entity.Payment) *PaymentOutput {
	return &PaymentOutput{
		ID:               p.ID,
		OwnershipShareID: p.OwnershipShareID,
		PaymentDate:      p.PaymentDate,
		Amount:           p.Amount,
		PaymentMethod:    string(p.PaymentMethod),
		PaymentPurpose:   p.PaymentPurpose,
		PeriodMonth:      p.PeriodMonth,
		PeriodYear:       p.PeriodYear,
		ReceiptNumber:    p.ReceiptNumber,
		Notes:            p.Notes,
		ApprovedBy:       p.ApprovedBy,
		ApprovedAt:       p.ApprovedAt,
		IsApproved:       p.IsApproved(),
		MethodName:       p.PaymentMethod.GetDisplayName(),
	}
}
