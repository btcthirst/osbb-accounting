// application/usecase/owner/owner_output.go
package owner

import (
	"fmt"
	"osbb-accounting/domain/entity"
)

// OwnerOutput - результат операцій з власником.
type OwnerOutput struct {
	ID                int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
	IsActive          bool
	FullName          string
	ShortName         string
}

// mapOwnerToOutput - helper function для конвертації entity в output.
func mapOwnerToOutput(o *entity.Owner) *OwnerOutput {
	return &OwnerOutput{
		ID:                o.ID,
		FirstName:         o.FirstName,
		LastName:          o.LastName,
		MiddleName:        o.MiddleName,
		Phone:             o.Phone,
		Email:             o.Email,
		TaxNumber:         o.TaxNumber,
		PassportSeries:    o.PassportSeries,
		PassportNumber:    o.PassportNumber,
		RegisteredAddress: o.RegisteredAddress,
		ActualAddress:     o.ActualAddress,
		Notes:             o.Notes,
		IsActive:          o.IsActive,
		FullName:          fmt.Sprintf("%s %s", o.LastName, o.FirstName),
		ShortName:         fmt.Sprintf("%s %s.", o.LastName, string([]rune(o.FirstName)[0])),
	}
}
