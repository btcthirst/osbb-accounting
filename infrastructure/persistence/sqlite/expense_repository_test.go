package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExpenseDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER NOT NULL,
		contractor_id INTEGER,
		expense_date INTEGER NOT NULL,
		amount REAL NOT NULL,
		description TEXT NOT NULL,
		document_type TEXT,
		document_number TEXT,
		document_date INTEGER,
		payment_status TEXT NOT NULL,
		paid_amount REAL NOT NULL,
		payment_date INTEGER,
		notes TEXT,
		approved_by INTEGER,
		approved_at INTEGER,
		deleted_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestExpenseRepository_Create(t *testing.T) {
	db := setupExpenseDB(t)
	defer db.Close()

	repo := sqlite.NewExpenseRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expense, err := entity.NewExpense(1, nil, time.Now(), 100.0, "Test Expense")
		require.NoError(t, err)

		err = repo.Create(ctx, expense)
		require.NoError(t, err)
		assert.NotZero(t, expense.ID)
	})
}

func TestExpenseRepository_List(t *testing.T) {
	db := setupExpenseDB(t)
	defer db.Close()

	repo := sqlite.NewExpenseRepository(db)
	ctx := context.Background()

	exp1, _ := entity.NewExpense(1, nil, time.Now(), 100.0, "Expense 1")
	repo.Create(ctx, exp1)
	exp2, _ := entity.NewExpense(2, nil, time.Now(), 200.0, "Expense 2")
	repo.Create(ctx, exp2)

	t.Run("List All", func(t *testing.T) {
		filter := repository.ExpenseFilter{}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Filter by Category", func(t *testing.T) {
		catID := int64(1)
		filter := repository.ExpenseFilter{CategoryID: &catID}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Expense 1", list[0].Description)
	})
}

func TestExpenseRepository_GetStatistics(t *testing.T) {
	db := setupExpenseDB(t)
	defer db.Close()

	repo := sqlite.NewExpenseRepository(db)
	ctx := context.Background()

	exp1, _ := entity.NewExpense(1, nil, time.Now(), 100.0, "Expense 1")
	exp1.PaymentStatus = entity.PaymentStatusPaid
	exp1.PaidAmount = 100.0
	repo.Create(ctx, exp1)

	exp2, _ := entity.NewExpense(1, nil, time.Now(), 50.0, "Expense 2")
	repo.Create(ctx, exp2)

	t.Run("Stats", func(t *testing.T) {
		filter := repository.ExpenseFilter{}
		stats, err := repo.GetStatistics(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), stats.TotalExpenses)
		assert.Equal(t, 150.0, stats.TotalAmount)
		assert.Equal(t, 100.0, stats.PaidAmount)
	})
}
