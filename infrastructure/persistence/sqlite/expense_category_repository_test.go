package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"osbb-accounting/domain/entity"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCategoryDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE expense_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT UNIQUE,
		parent_id INTEGER,
		category_type TEXT NOT NULL,
		description TEXT,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (parent_id) REFERENCES expense_categories(id)
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestExpenseCategoryRepository_Create(t *testing.T) {
	db := setupCategoryDB(t)
	defer db.Close()

	repo := sqlite.NewExpenseCategoryRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		cat, err := entity.NewExpenseCategory("Utilities", entity.CategoryTypeUtility, nil, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, cat)
		require.NoError(t, err)
		assert.NotZero(t, cat.ID)
	})

	t.Run("Duplicate Code", func(t *testing.T) {
		code := "UTIL"
		c1, _ := entity.NewExpenseCategory("Cat1", entity.CategoryTypeUtility, nil, &code, nil)
		repo.Create(ctx, c1)

		c2, _ := entity.NewExpenseCategory("Cat2", entity.CategoryTypeUtility, nil, &code, nil)
		err := repo.Create(ctx, c2)
		assert.Error(t, err)
	})
}

func TestExpenseCategoryRepository_Hierarchy(t *testing.T) {
	db := setupCategoryDB(t)
	defer db.Close()

	repo := sqlite.NewExpenseCategoryRepository(db)
	ctx := context.Background()

	parent, _ := entity.NewExpenseCategory("Parent", entity.CategoryTypeUtility, nil, nil, nil)
	repo.Create(ctx, parent)

	child, _ := entity.NewExpenseCategory("Child", entity.CategoryTypeUtility, &parent.ID, nil, nil)
	err := repo.Create(ctx, child)
	require.NoError(t, err)

	t.Run("Get Children", func(t *testing.T) {
		children, err := repo.GetChildren(ctx, parent.ID)
		require.NoError(t, err)
		assert.Len(t, children, 1)
		assert.Equal(t, "Child", children[0].Name)
	})

	t.Run("Get Depth", func(t *testing.T) {
		depth, err := repo.GetDepth(ctx, child.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, depth)
	})
}
