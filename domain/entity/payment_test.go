package entity_test

import (
	"testing"
	"time"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		payment, err := entity.NewPayment(
			1,
			100.0,
			entity.PaymentMethodCash,
			"Test Payment",
			time.Now(),
		)
		require.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, 100.0, payment.Amount)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name    string
			amount  float64
			method  entity.PaymentMethod
			wantErr error
		}{
			{"Invalid Amount", -10, entity.PaymentMethodCash, entity.ErrPaymentAmountInvalid},
			{"Invalid Method", 100, "invalid", entity.ErrPaymentMethodInvalid},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewPayment(1, tt.amount, tt.method, "Test", time.Now())
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestPayment_SetPeriod(t *testing.T) {
	payment, _ := entity.NewPayment(1, 100.0, entity.PaymentMethodCash, "Test", time.Now())

	t.Run("Success", func(t *testing.T) {
		err := payment.SetPeriod(1, 2024)
		require.NoError(t, err)
		assert.Equal(t, 1, *payment.PeriodMonth)
	})

	t.Run("Invalid Period", func(t *testing.T) {
		err := payment.SetPeriod(13, 2024)
		assert.ErrorIs(t, err, entity.ErrPaymentPeriodInvalid)
	})
}

func TestPayment_Approval(t *testing.T) {
	payment, _ := entity.NewPayment(1, 100.0, entity.PaymentMethodCash, "Test", time.Now())

	t.Run("Approve", func(t *testing.T) {
		err := payment.Approve(1)
		require.NoError(t, err)
		assert.True(t, payment.IsApproved())
	})

	t.Run("Already Approved", func(t *testing.T) {
		err := payment.Approve(1)
		assert.ErrorIs(t, err, entity.ErrPaymentAlreadyApproved)
	})
}
