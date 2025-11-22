package entity_test

import (
	"testing"
	"time"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpense(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expense, err := entity.NewExpense(1, nil, time.Now(), 100.0, "Test Expense")
		require.NoError(t, err)
		assert.NotNil(t, expense)
		assert.Equal(t, entity.PaymentStatusPending, expense.PaymentStatus)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		_, err := entity.NewExpense(0, nil, time.Now(), 100.0, "Test")
		assert.ErrorIs(t, err, entity.ErrExpenseCategoryIDRequired)

		_, err = entity.NewExpense(1, nil, time.Now(), -10.0, "Test")
		assert.ErrorIs(t, err, entity.ErrExpenseAmountInvalid)
	})
}

func TestExpense_RecordPayment(t *testing.T) {
	expense, _ := entity.NewExpense(1, nil, time.Now(), 100.0, "Test")

	t.Run("Partial Payment", func(t *testing.T) {
		err := expense.RecordPayment(50.0, time.Now())
		require.NoError(t, err)
		assert.Equal(t, 50.0, expense.PaidAmount)
		assert.Equal(t, entity.PaymentStatusPartiallyPaid, expense.PaymentStatus)
	})

	t.Run("Full Payment", func(t *testing.T) {
		err := expense.RecordPayment(50.0, time.Now())
		require.NoError(t, err)
		assert.Equal(t, 100.0, expense.PaidAmount)
		assert.Equal(t, entity.PaymentStatusPaid, expense.PaymentStatus)
	})

	t.Run("Over Payment", func(t *testing.T) {
		err := expense.RecordPayment(10.0, time.Now())
		assert.Error(t, err)
	})
}

func TestExpense_Approval(t *testing.T) {
	expense, _ := entity.NewExpense(1, nil, time.Now(), 100.0, "Test")

	t.Run("Approve", func(t *testing.T) {
		err := expense.Approve(1)
		require.NoError(t, err)
		assert.True(t, expense.IsApproved())
	})

	t.Run("Already Approved", func(t *testing.T) {
		err := expense.Approve(1)
		assert.ErrorIs(t, err, entity.ErrExpenseAlreadyApproved)
	})

	t.Run("Unapprove", func(t *testing.T) {
		err := expense.Unapprove()
		require.NoError(t, err)
		assert.False(t, expense.IsApproved())
	})
}
