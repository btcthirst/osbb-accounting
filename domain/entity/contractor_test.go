package entity_test

import (
	"testing"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContractor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		contractor, err := entity.NewContractor(
			"Test Contractor",
			entity.ContractorTypeService,
			nil, nil, nil, nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, contractor)
		assert.True(t, contractor.IsActive)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name    string
			cName   string
			cType   entity.ContractorType
			edrpou  string
			wantErr error
		}{
			{"Empty Name", "", entity.ContractorTypeService, "", entity.ErrContractorNameRequired},
			{"Short Name", "A", entity.ContractorTypeService, "", entity.ErrContractorNameTooShort},
			{"Invalid Type", "Test", "invalid", "", entity.ErrContractorTypeInvalid},
			{"Invalid EDRPOU", "Test", entity.ContractorTypeService, "123", entity.ErrContractorEDRPOUInvalidFormat},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var edrpou *string
				if tt.edrpou != "" {
					edrpou = &tt.edrpou
				}
				_, err := entity.NewContractor(tt.cName, tt.cType, edrpou, nil, nil, nil)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestContractor_UpdateBankDetails(t *testing.T) {
	contractor, _ := entity.NewContractor("Test", entity.ContractorTypeService, nil, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		iban := "UA123456789012345678901234567" // 29 chars
		mfo := "123456"
		err := contractor.UpdateBankDetails(&iban, nil, &mfo)
		require.NoError(t, err)
		assert.Equal(t, iban, *contractor.BankAccount)
	})

	t.Run("Invalid IBAN", func(t *testing.T) {
		iban := "short"
		err := contractor.UpdateBankDetails(&iban, nil, nil)
		assert.ErrorIs(t, err, entity.ErrContractorBankAccountInvalid)
	})
}
