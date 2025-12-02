// application/usecase/charge/charge_output.go
package charge

import (
	"time"

	"osbb-accounting/domain/entity"
)

// ChargeOutput - результат операцій з нарахуванням.
type ChargeOutput struct {
	ID               int64
	OwnershipShareID int64
	ChargeType       string
	ChargeDate       time.Time
	PeriodMonth      int
	PeriodYear       int
	Amount           float64
	Tariff           *float64
	Quantity         *float64
	Description      *string
	Notes            *string
	PeriodDisplay    string
	AmountDisplay    string
	TypeName         string
}

// mapChargeToOutput - helper function для конвертації entity в output.
func mapChargeToOutput(c *entity.Charge) *ChargeOutput {
	return &ChargeOutput{
		ID:               c.ID,
		OwnershipShareID: c.OwnershipShareID,
		ChargeType:       string(c.ChargeType),
		ChargeDate:       c.ChargeDate,
		PeriodMonth:      c.PeriodMonth,
		PeriodYear:       c.PeriodYear,
		Amount:           c.Amount,
		Tariff:           c.Tariff,
		Quantity:         c.Quantity,
		Description:      c.Description,
		Notes:            c.Notes,
		PeriodDisplay:    c.GetPeriodDisplay(),
		AmountDisplay:    c.GetAmountDisplay(),
		TypeName:         c.ChargeType.GetDisplayName(),
	}
}
