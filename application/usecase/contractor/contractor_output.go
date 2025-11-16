// application/usecase/contractor/contractor_output.go
package contractor

import (
	"time"

	"osbb-accounting/domain/entity"
)

// ContractorOutput - результат операцій з контрагентом.
type ContractorOutput struct {
	ID              int64
	Name            string
	EDRPOU          *string
	ContractorType  entity.ContractorType
	TypeDisplayName string
	ContactPerson   *string
	Phone           *string
	Email           *string
	Address         *string
	BankAccount     *string
	BankName        *string
	BankMFO         *string
	ContractNumber  *string
	ContractDate    *time.Time
	Notes           *string
	IsActive        bool
	DisplayName     string
	ContactInfo     string
	HasBankDetails  bool
	HasContract     bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// mapContractorToOutput - helper для конвертації entity в output.
func mapContractorToOutput(c *entity.Contractor) *ContractorOutput {
	return &ContractorOutput{
		ID:              c.ID,
		Name:            c.Name,
		EDRPOU:          c.EDRPOU,
		ContractorType:  c.ContractorType,
		TypeDisplayName: c.ContractorType.GetDisplayName(),
		ContactPerson:   c.ContactPerson,
		Phone:           c.Phone,
		Email:           c.Email,
		Address:         c.Address,
		BankAccount:     c.BankAccount,
		BankName:        c.BankName,
		BankMFO:         c.BankMFO,
		ContractNumber:  c.ContractNumber,
		ContractDate:    c.ContractDate,
		Notes:           c.Notes,
		IsActive:        c.IsActive,
		DisplayName:     c.GetDisplayName(),
		ContactInfo:     c.GetContactInfo(),
		HasBankDetails:  c.HasBankDetails(),
		HasContract:     c.HasContract(),
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
