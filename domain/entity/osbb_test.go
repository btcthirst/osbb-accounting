package entity_test

import (
	"testing"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOSBB(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		osbb, err := entity.NewOSBB(
			"My OSBB",
			"12345678",
			"Legal Addr",
			"Chairman",
			nil, nil, nil, nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, osbb)
		assert.Equal(t, "My OSBB", osbb.Name)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name     string
			osbbName string
			edrpou   string
			wantErr  error
		}{
			{"Empty Name", "", "12345678", entity.ErrOSBBNameRequired},
			{"Short Name", "A", "12345678", entity.ErrOSBBNameTooShort},
			{"Invalid EDRPOU", "Test", "123", entity.ErrOSBBEDRPOUInvalidFormat},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewOSBB(tt.osbbName, tt.edrpou, "Addr", "Chair", nil, nil, nil, nil)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestOSBB_Update(t *testing.T) {
	osbb, _ := entity.NewOSBB("Test", "12345678", "Addr", "Chair", nil, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		err := osbb.Update("New Name", "New Addr", "New Chair", nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "New Name", osbb.Name)
	})

	t.Run("Validation Error", func(t *testing.T) {
		err := osbb.Update("", "Addr", "Chair", nil, nil, nil, nil)
		assert.ErrorIs(t, err, entity.ErrOSBBNameRequired)
	})
}
