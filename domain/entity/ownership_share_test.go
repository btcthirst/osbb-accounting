package entity_test

import (
	"testing"
	"time"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOwnershipShare(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		share, err := entity.NewOwnershipShare(1, 1, 1, 2, entity.OwnershipTypeShared, time.Now())
		require.NoError(t, err)
		assert.NotNil(t, share)
		assert.Equal(t, 50.0, share.GetSharePercentage())
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name    string
			num     int
			denom   int
			wantErr error
		}{
			{"Invalid Numerator", 0, 2, entity.ErrOwnershipShareNumeratorInvalid},
			{"Invalid Denominator", 2, 1, entity.ErrOwnershipShareDenominatorInvalid},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewOwnershipShare(1, 1, tt.num, tt.denom, entity.OwnershipTypeShared, time.Now())
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestOwnershipShare_IsCurrentlyActive(t *testing.T) {
	now := time.Now()

	t.Run("Active", func(t *testing.T) {
		share, _ := entity.NewOwnershipShare(1, 1, 1, 1, entity.OwnershipTypeFull, now.AddDate(0, -1, 0))
		assert.True(t, share.IsCurrentlyActive())
	})

	t.Run("Future Start", func(t *testing.T) {
		share, _ := entity.NewOwnershipShare(1, 1, 1, 1, entity.OwnershipTypeFull, now.AddDate(0, 1, 0))
		assert.False(t, share.IsCurrentlyActive())
	})

	t.Run("Terminated", func(t *testing.T) {
		share, _ := entity.NewOwnershipShare(1, 1, 1, 1, entity.OwnershipTypeFull, now.AddDate(0, -2, 0))
		err := share.TerminateOwnership(now.AddDate(0, -1, 0))
		require.NoError(t, err)
		assert.False(t, share.IsCurrentlyActive())
	})
}
