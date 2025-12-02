package expense_category

import (
	"time"

	"osbb-accounting/domain/entity"
)

// ============================================================================
// OUTPUT TYPES
// ============================================================================

// ExpenseCategoryOutput - результат операцій з категорією витрат.
type ExpenseCategoryOutput struct {
	ID              int64
	Name            string
	Code            *string
	ParentID        *int64
	CategoryType    entity.CategoryType
	TypeDisplayName string
	Description     *string
	IsActive        bool
	Level           int
	Path            string
	DisplayName     string
	HasChildren     bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ExpenseCategoryTreeOutput - категорія з дочірніми елементами.
type ExpenseCategoryTreeOutput struct {
	ExpenseCategoryOutput
	Children []*ExpenseCategoryTreeOutput
}
