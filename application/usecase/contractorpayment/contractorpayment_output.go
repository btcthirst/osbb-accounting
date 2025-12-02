// application/usecase/contractorpayment/contractorpayment_output.go
package contractorpayment

import (
	"time"

	"osbb-accounting/domain/entity"
)

// ContractorPaymentOutput - результат операцій з платежем контрагента.
type ContractorPaymentOutput struct {
	ID            int64
	ContractorID  int64
	PaymentDate   time.Time
	Amount        float64
	PaymentMethod string
	Purpose       string
	PeriodMonth   *int
	PeriodYear    *int
	ReceiptNumber *string
	Notes         *string
	MethodName    string
}

// DeleteContractorPaymentOutput - результат видалення.
type DeleteContractorPaymentOutput struct {
	Success bool
	Message string
}

// mapContractorPaymentToOutput - helper function для конвертації entity в output.
func mapContractorPaymentToOutput(p *entity.ContractorPayment) *ContractorPaymentOutput {
	return &ContractorPaymentOutput{
		ID:            p.ID,
		ContractorID:  p.ContractorID,
		PaymentDate:   p.PaymentDate,
		Amount:        p.Amount,
		PaymentMethod: string(p.PaymentMethod),
		Purpose:       p.Purpose,
		PeriodMonth:   p.PeriodMonth,
		PeriodYear:    p.PeriodYear,
		ReceiptNumber: p.ReceiptNumber,
		Notes:         p.Notes,
		MethodName:    p.PaymentMethod.GetDisplayName(),
	}
}
