// infrastructure/persistence/sqlite/expense_category_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ExpenseCategoryRepository реалізує repository.ExpenseCategoryRepository для SQLite.
type ExpenseCategoryRepository struct {
	db *sql.DB
}

// NewExpenseCategoryRepository створює новий ExpenseCategoryRepository.
func NewExpenseCategoryRepository(db *sql.DB) *ExpenseCategoryRepository {
	return &ExpenseCategoryRepository{db: db}
}

// Create створює нову категорію витрат.
func (r *ExpenseCategoryRepository) Create(ctx context.Context, category *entity.ExpenseCategory) error {
	// Перевірка глибини вкладеності
	if category.ParentID != nil {
		depth, err := r.GetDepth(ctx, *category.ParentID)
		if err != nil {
			return fmt.Errorf("failed to check parent depth: %w", err)
		}
		if depth >= entity.MaxCategoryDepth-1 {
			return entity.ErrCategoryMaxDepthExceeded
		}
	}

	query := `
		INSERT INTO expense_categories (
			name, code, parent_id, category_type, description, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Code,
		category.ParentID,
		string(category.CategoryType),
		category.Description,
		boolToInt(category.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "code") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"category with this code already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "code")
			}
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return entity.ErrCategoryParentNotFound
		}
		return fmt.Errorf("failed to create expense category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	category.ID = id
	category.CreatedAt = time.Unix(now, 0)
	category.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує категорію за ID.
func (r *ExpenseCategoryRepository) GetByID(ctx context.Context, id int64) (*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanCategory(ctx, query, id)
}

// GetByCode отримує категорію за кодом.
func (r *ExpenseCategoryRepository) GetByCode(ctx context.Context, code string) (*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE code = ? AND deleted_at IS NULL
	`

	return r.scanCategory(ctx, query, code)
}

// Update оновлює дані категорії.
func (r *ExpenseCategoryRepository) Update(ctx context.Context, category *entity.ExpenseCategory) error {
	query := `
		UPDATE expense_categories
		SET name = ?,
		    code = ?,
		    category_type = ?,
		    description = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Code,
		string(category.CategoryType),
		category.Description,
		boolToInt(category.IsActive),
		now,
		category.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"category with this code already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "code")
		}
		return fmt.Errorf("failed to update expense category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	category.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення категорії.
func (r *ExpenseCategoryRepository) SoftDelete(ctx context.Context, id int64) error {
	// Перевірка чи має категорія дочірні елементи
	hasChildren, err := r.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if hasChildren {
		return entity.ErrCategoryHasChildren
	}

	query := `
		UPDATE expense_categories
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete expense category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// Restore відновлює видалену категорію.
func (r *ExpenseCategoryRepository) Restore(ctx context.Context, id int64) error {
	query := `
		UPDATE expense_categories
		SET deleted_at = NULL, is_active = 1, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore expense category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// List отримує список категорій з фільтрацією.
func (r *ExpenseCategoryRepository) List(ctx context.Context, filter repository.ExpenseCategoryFilter) ([]*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.SearchQuery != "" {
		query += " AND (name LIKE ? OR code LIKE ?)"
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filter.CategoryType != nil {
		query += " AND category_type = ?"
		args = append(args, string(*filter.CategoryType))
	}

	if filter.OnlyRoot {
		query += " AND parent_id IS NULL"
	} else if filter.ParentID != nil {
		query += " AND parent_id = ?"
		args = append(args, *filter.ParentID)
	}

	// Сортування
	orderBy := "name"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "name":
			orderBy = "name"
		case "code":
			orderBy = "code"
		case "type":
			orderBy = "category_type, name"
		case "created_at":
			orderBy = "created_at"
		}
	}

	if filter.OrderDesc {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderBy)
	} else {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderBy)
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list expense categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// Count отримує загальну кількість категорій.
func (r *ExpenseCategoryRepository) Count(ctx context.Context, filter repository.ExpenseCategoryFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM expense_categories WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.SearchQuery != "" {
		query += " AND (name LIKE ? OR code LIKE ?)"
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filter.CategoryType != nil {
		query += " AND category_type = ?"
		args = append(args, string(*filter.CategoryType))
	}

	if filter.OnlyRoot {
		query += " AND parent_id IS NULL"
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count expense categories: %w", err)
	}

	return count, nil
}

// GetRootCategories отримує всі кореневі категорії.
func (r *ExpenseCategoryRepository) GetRootCategories(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY category_type, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// GetChildren отримує всі дочірні категорії.
func (r *ExpenseCategoryRepository) GetChildren(ctx context.Context, parentID int64) ([]*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE parent_id = ? AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// GetTree отримує повне дерево категорій.
func (r *ExpenseCategoryRepository) GetTree(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	// Отримуємо всі категорії
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE deleted_at IS NULL
		ORDER BY category_type, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all categories: %w", err)
	}
	defer rows.Close()

	allCategories, err := r.scanCategories(rows)
	if err != nil {
		return nil, err
	}

	// Будуємо дерево
	categoryMap := make(map[int64]*entity.ExpenseCategory)
	roots := make([]*entity.ExpenseCategory, 0)

	// Створюємо мапу всіх категорій
	for _, cat := range allCategories {
		categoryMap[cat.ID] = cat
		cat.Children = make([]*entity.ExpenseCategory, 0)
	}

	// Будуємо дерево
	for _, cat := range allCategories {
		if cat.ParentID == nil {
			// Коренева категорія
			roots = append(roots, cat)
		} else {
			// Дочірня категорія
			if parent, ok := categoryMap[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, cat)
				cat.Level = parent.Level + 1
			}
		}
	}

	return roots, nil
}

// GetAncestors отримує всіх предків категорії.
func (r *ExpenseCategoryRepository) GetAncestors(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error) {
	// Рекурсивний запит для отримання предків
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, code, parent_id, category_type, description, is_active,
			       deleted_at, created_at, updated_at, 0 as level
			FROM expense_categories
			WHERE id = ?
			
			UNION ALL
			
			SELECT c.id, c.name, c.code, c.parent_id, c.category_type, c.description, c.is_active,
			       c.deleted_at, c.created_at, c.updated_at, a.level + 1
			FROM expense_categories c
			INNER JOIN ancestors a ON c.id = a.parent_id
			WHERE c.deleted_at IS NULL
		)
		SELECT id, name, code, parent_id, category_type, description, is_active,
		       deleted_at, created_at, updated_at
		FROM ancestors
		WHERE id != ?
		ORDER BY level DESC
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// GetDescendants отримує всіх нащадків категорії.
func (r *ExpenseCategoryRepository) GetDescendants(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error) {
	// Рекурсивний запит для отримання нащадків
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, name, code, parent_id, category_type, description, is_active,
			       deleted_at, created_at, updated_at, 0 as level
			FROM expense_categories
			WHERE parent_id = ? AND deleted_at IS NULL
			
			UNION ALL
			
			SELECT c.id, c.name, c.code, c.parent_id, c.category_type, c.description, c.is_active,
			       c.deleted_at, c.created_at, c.updated_at, d.level + 1
			FROM expense_categories c
			INNER JOIN descendants d ON c.parent_id = d.id
			WHERE c.deleted_at IS NULL
		)
		SELECT id, name, code, parent_id, category_type, description, is_active,
		       deleted_at, created_at, updated_at
		FROM descendants
		ORDER BY level, name
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// HasChildren перевіряє чи має категорія дочірні елементи.
func (r *ExpenseCategoryRepository) HasChildren(ctx context.Context, categoryID int64) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM expense_categories 
		WHERE parent_id = ? AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check children: %w", err)
	}

	return count > 0, nil
}

// ExistsByCode перевіряє чи існує категорія з таким кодом.
func (r *ExpenseCategoryRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT COUNT(*) FROM expense_categories WHERE code = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, code).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return count > 0, nil
}

// GetDepth отримує глибину категорії в дереві.
func (r *ExpenseCategoryRepository) GetDepth(ctx context.Context, categoryID int64) (int, error) {
	query := `
		WITH RECURSIVE depth_calc AS (
			SELECT id, parent_id, 0 as depth
			FROM expense_categories
			WHERE id = ?
			
			UNION ALL
			
			SELECT c.id, c.parent_id, d.depth + 1
			FROM expense_categories c
			INNER JOIN depth_calc d ON c.id = d.parent_id
		)
		SELECT MAX(depth) FROM depth_calc
	`

	var depth int
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&depth)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate depth: %w", err)
	}

	return depth, nil
}

// Move переміщує категорію до іншого батька.
func (r *ExpenseCategoryRepository) Move(ctx context.Context, categoryID int64, newParentID *int64) error {
	// Перевірка циклічного посилання
	if newParentID != nil {
		if *newParentID == categoryID {
			return entity.ErrCategoryCannotBeOwnParent
		}

		// Перевірка що newParent не є нащадком categoryID
		descendants, err := r.GetDescendants(ctx, categoryID)
		if err != nil {
			return fmt.Errorf("failed to check descendants: %w", err)
		}

		for _, desc := range descendants {
			if desc.ID == *newParentID {
				return entity.ErrCategoryCircularReference
			}
		}

		// Перевірка глибини

		depth, err := r.GetDepth(ctx, *newParentID)
		if err != nil {
			return fmt.Errorf("failed to check parent depth: %w", err)
		}
		if depth >= entity.MaxCategoryDepth-1 {
			return entity.ErrCategoryMaxDepthExceeded
		}

	}

	query := `
		UPDATE expense_categories
		SET parent_id = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, newParentID, now, categoryID)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return entity.ErrCategoryParentNotFound
		}
		return fmt.Errorf("failed to move category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// GetByType отримує всі категорії певного типу.
func (r *ExpenseCategoryRepository) GetByType(ctx context.Context, categoryType entity.CategoryType) ([]*entity.ExpenseCategory, error) {
	query := `
		SELECT 
			id, name, code, parent_id, category_type, description, is_active,
			deleted_at, created_at, updated_at
		FROM expense_categories
		WHERE category_type = ? AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, string(categoryType))
	if err != nil {
		return nil, fmt.Errorf("failed to get categories by type: %w", err)
	}
	defer rows.Close()

	return r.scanCategories(rows)
}

// ============================================================================
// Helper функції
// ============================================================================

// scanCategory - helper для сканування однієї категорії.
func (r *ExpenseCategoryRepository) scanCategory(ctx context.Context, query string, args ...interface{}) (*entity.ExpenseCategory, error) {
	category := &entity.ExpenseCategory{}
	var code, description sql.NullString
	var parentID sql.NullInt64
	var deletedAt, createdAt, updatedAt sql.NullInt64
	var categoryType string
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&category.ID,
		&category.Name,
		&code,
		&parentID,
		&categoryType,
		&description,
		&isActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan expense category: %w", err)
	}

	category.Code = nullStringToPtr(code)
	category.Description = nullStringToPtr(description)
	if parentID.Valid {
		pid := parentID.Int64
		category.ParentID = &pid
	}
	category.CategoryType = entity.CategoryType(categoryType)
	category.IsActive = intToBool(isActive)
	category.DeletedAt = nullInt64ToTimePtr(deletedAt)
	category.CreatedAt = time.Unix(createdAt.Int64, 0)
	category.UpdatedAt = time.Unix(updatedAt.Int64, 0)
	category.Children = make([]*entity.ExpenseCategory, 0)

	return category, nil
}

// scanCategories - helper для сканування списку категорій.
func (r *ExpenseCategoryRepository) scanCategories(rows *sql.Rows) ([]*entity.ExpenseCategory, error) {
	categories := make([]*entity.ExpenseCategory, 0)

	for rows.Next() {
		category := &entity.ExpenseCategory{}
		var code, description sql.NullString
		var parentID sql.NullInt64
		var deletedAt, createdAt, updatedAt sql.NullInt64
		var categoryType string
		var isActive int

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&code,
			&parentID,
			&categoryType,
			&description,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense category: %w", err)
		}

		category.Code = nullStringToPtr(code)
		category.Description = nullStringToPtr(description)
		if parentID.Valid {
			pid := parentID.Int64
			category.ParentID = &pid
		}
		category.CategoryType = entity.CategoryType(categoryType)
		category.IsActive = intToBool(isActive)
		category.DeletedAt = nullInt64ToTimePtr(deletedAt)
		category.CreatedAt = time.Unix(createdAt.Int64, 0)
		category.UpdatedAt = time.Unix(updatedAt.Int64, 0)
		category.Children = make([]*entity.ExpenseCategory, 0)

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expense categories: %w", err)
	}

	return categories, nil
}
