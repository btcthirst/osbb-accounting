// infrastructure/persistence/sqlite/expense_repository.go
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

// ExpenseRepository реалізує repository.ExpenseRepository для SQLite.
type ExpenseRepository struct {
	db *sql.DB
}

// NewExpenseRepository створює новий ExpenseRepository.
func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create створює нову витрату.
func (r *ExpenseRepository) Create(ctx context.Context, expense *entity.Expense) error {
	query := `
		INSERT INTO expenses (
			category_id, contractor_id, expense_date, amount, description,
			document_type, document_number, document_date,
			payment_status, paid_amount, payment_date,
			notes, approved_by, approved_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		expense.CategoryID,
		expense.ContractorID,
		expense.ExpenseDate.Unix(),
		expense.Amount,
		expense.Description,
		expense.DocumentType,
		expense.DocumentNumber,
		timeToNullInt64(expense.DocumentDate),
		string(expense.PaymentStatus),
		expense.PaidAmount,
		timeToNullInt64(expense.PaymentDate),
		expense.Notes,
		expense.ApprovedBy,
		timeToNullInt64(expense.ApprovedAt),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domainErrors.ErrForeignKeyViolation
		}
		return fmt.Errorf("failed to create expense: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	expense.ID = id
	expense.CreatedAt = time.Unix(now, 0)
	expense.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує витрату за ID.
func (r *ExpenseRepository) GetByID(ctx context.Context, id int64) (*entity.Expense, error) {
	query := `
		SELECT 
			id, category_id, contractor_id, expense_date, amount, description,
			document_type, document_number, document_date,
			payment_status, paid_amount, payment_date,
			notes, approved_by, approved_at,
			deleted_at, created_at, updated_at
		FROM expenses
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanExpense(ctx, query, id)
}

// Update оновлює дані витрати.
func (r *ExpenseRepository) Update(ctx context.Context, expense *entity.Expense) error {
	query := `
		UPDATE expenses
		SET category_id = ?,
		    contractor_id = ?,
		    expense_date = ?,
		    amount = ?,
		    description = ?,
		    document_type = ?,
		    document_number = ?,
		    document_date = ?,
		    payment_status = ?,
		    paid_amount = ?,
		    payment_date = ?,
		    notes = ?,
		    approved_by = ?,
		    approved_at = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		expense.CategoryID,
		expense.ContractorID,
		expense.ExpenseDate.Unix(),
		expense.Amount,
		expense.Description,
		expense.DocumentType,
		expense.DocumentNumber,
		timeToNullInt64(expense.DocumentDate),
		string(expense.PaymentStatus),
		expense.PaidAmount,
		timeToNullInt64(expense.PaymentDate),
		expense.Notes,
		expense.ApprovedBy,
		timeToNullInt64(expense.ApprovedAt),
		now,
		expense.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domainErrors.ErrForeignKeyViolation
		}
		return fmt.Errorf("failed to update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	expense.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення витрати.
func (r *ExpenseRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE expenses
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete expense: %w", err)
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

// Restore відновлює видалену витрату.
func (r *ExpenseRepository) Restore(ctx context.Context, id int64) error {
	query := `
		UPDATE expenses
		SET deleted_at = NULL, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore expense: %w", err)
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

// List отримує список витрат з фільтрацією.
func (r *ExpenseRepository) List(ctx context.Context, filter repository.ExpenseFilter) ([]*entity.Expense, error) {
	query := `
		SELECT 
			id, category_id, contractor_id, expense_date, amount, description,
			document_type, document_number, document_date,
			payment_status, paid_amount, payment_date,
			notes, approved_by, approved_at,
			deleted_at, created_at, updated_at
		FROM expenses
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	// Фільтр по видаленим
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	// Пошук
	if filter.SearchQuery != "" {
		query += " AND (description LIKE ? OR document_number LIKE ?)"
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Фільтр по категорії
	if filter.CategoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	// Фільтр по контрагенту
	if filter.ContractorID != nil {
		query += " AND contractor_id = ?"
		args = append(args, *filter.ContractorID)
	}

	// Фільтр по статусу оплати
	if filter.PaymentStatus != nil {
		query += " AND payment_status = ?"
		args = append(args, string(*filter.PaymentStatus))
	}

	// Фільтр по затвердженню
	if filter.IsApproved != nil {
		if *filter.IsApproved {
			query += " AND approved_by IS NOT NULL"
		} else {
			query += " AND approved_by IS NULL"
		}
	}

	// Фільтр по користувачу, який затвердив
	if filter.ApprovedBy != nil {
		query += " AND approved_by = ?"
		args = append(args, *filter.ApprovedBy)
	}

	// Фільтр по даті
	if filter.StartDate != nil {
		query += " AND expense_date >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND expense_date <= ?"
		args = append(args, *filter.EndDate)
	}

	// Фільтр по сумі
	if filter.MinAmount != nil {
		query += " AND amount >= ?"
		args = append(args, *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		query += " AND amount <= ?"
		args = append(args, *filter.MaxAmount)
	}

	// Фільтр по типу документа
	if filter.DocumentType != nil {
		query += " AND document_type = ?"
		args = append(args, *filter.DocumentType)
	}

	// Сортування
	orderBy := "expense_date DESC"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "expense_date":
			orderBy = "expense_date"
		case "amount":
			orderBy = "amount"
		case "created_at":
			orderBy = "created_at"
		case "updated_at":
			orderBy = "updated_at"
		}
		if filter.OrderDesc {
			orderBy += " DESC"
		} else {
			orderBy += " ASC"
		}
	}
	query += " ORDER BY " + orderBy

	// Pagination
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
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	return r.scanExpenses(rows)
}

// Count отримує загальну кількість витрат.
func (r *ExpenseRepository) Count(ctx context.Context, filter repository.ExpenseFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM expenses WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.SearchQuery != "" {
		query += " AND (description LIKE ? OR document_number LIKE ?)"
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filter.CategoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	if filter.ContractorID != nil {
		query += " AND contractor_id = ?"
		args = append(args, *filter.ContractorID)
	}

	if filter.PaymentStatus != nil {
		query += " AND payment_status = ?"
		args = append(args, string(*filter.PaymentStatus))
	}

	if filter.StartDate != nil {
		query += " AND expense_date >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND expense_date <= ?"
		args = append(args, *filter.EndDate)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count expenses: %w", err)
	}

	return count, nil
}

// GetByCategoryID отримує всі витрати категорії.
func (r *ExpenseRepository) GetByCategoryID(ctx context.Context, categoryID int64) ([]*entity.Expense, error) {
	return r.List(ctx, repository.ExpenseFilter{
		CategoryID: &categoryID,
		OrderBy:    "expense_date",
		OrderDesc:  true,
	})
}

// GetByContractorID отримує всі витрати контрагента.
func (r *ExpenseRepository) GetByContractorID(ctx context.Context, contractorID int64) ([]*entity.Expense, error) {
	return r.List(ctx, repository.ExpenseFilter{
		ContractorID: &contractorID,
		OrderBy:      "expense_date",
		OrderDesc:    true,
	})
}

// GetByDateRange отримує витрати за період.
func (r *ExpenseRepository) GetByDateRange(ctx context.Context, startDate, endDate int64) ([]*entity.Expense, error) {
	return r.List(ctx, repository.ExpenseFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
		OrderBy:   "expense_date",
		OrderDesc: false,
	})
}

// GetStatistics отримує статистику по витратах.
func (r *ExpenseRepository) GetStatistics(ctx context.Context, filter repository.ExpenseFilter) (*repository.ExpenseStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_expenses,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(paid_amount), 0) as paid_amount,
			COALESCE(SUM(amount - paid_amount), 0) as unpaid_amount,
			COALESCE(AVG(amount), 0) as avg_amount,
			COALESCE(MIN(amount), 0) as min_amount,
			COALESCE(MAX(amount), 0) as max_amount,
			SUM(CASE WHEN payment_status = 'partially_paid' THEN 1 ELSE 0 END) as partially_paid,
			SUM(CASE WHEN payment_status = 'paid' THEN 1 ELSE 0 END) as fully_paid,
			SUM(CASE WHEN payment_status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN payment_status = 'cancelled' THEN 1 ELSE 0 END) as cancelled,
			SUM(CASE WHEN approved_by IS NOT NULL THEN 1 ELSE 0 END) as approved_count,
			SUM(CASE WHEN payment_status = 'partially_paid' THEN 1 ELSE 0 END) as partially_paid,
			SUM(CASE WHEN payment_status = 'paid' THEN 1 ELSE 0 END) as fully_paid,
			SUM(CASE WHEN payment_status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN payment_status = 'cancelled' THEN 1 ELSE 0 END) as cancelled,
			SUM(CASE WHEN approved_by IS NOT NULL THEN 1 ELSE 0 END) as approved_count,
			SUM(CASE WHEN approved_by IS NULL THEN 1 ELSE 0 END) as not_approved_count
		FROM expenses
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	// Застосовуємо ті ж фільтри, що і в List/Count для узгодженості даних
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.SearchQuery != "" {
		query += " AND (description LIKE ? OR document_number LIKE ?)"
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filter.CategoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	if filter.ContractorID != nil {
		query += " AND contractor_id = ?"
		args = append(args, *filter.ContractorID)
	}

	if filter.PaymentStatus != nil {
		query += " AND payment_status = ?"
		args = append(args, string(*filter.PaymentStatus))
	}

	if filter.StartDate != nil {
		query += " AND expense_date >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND expense_date <= ?"
		args = append(args, *filter.EndDate)
	}

	stats := &repository.ExpenseStatistics{}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalExpenses,
		&stats.TotalAmount,
		&stats.PaidAmount,
		&stats.UnpaidAmount,
		&stats.AverageAmount,
		&stats.MinAmount,
		&stats.MaxAmount,
		&stats.PartiallyPaid,
		&stats.FullyPaid,
		&stats.Pending,
		&stats.Cancelled,
		&stats.ApprovedCount,
		&stats.NotApprovedCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get expense statistics: %w", err)
	}

	// Розрахунок відсотка оплати (щоб уникнути ділення на нуль)
	if stats.TotalAmount > 0 {
		stats.PaymentRate = (stats.PaidAmount / stats.TotalAmount) * 100
	} else {
		stats.PaymentRate = 0
	}

	return stats, nil
}

// GetTotalByCategory отримує загальну суму витрат по категоріях.
func (r *ExpenseRepository) GetTotalByCategory(ctx context.Context, startDate, endDate int64) (map[int64]float64, error) {
	query := `
		SELECT category_id, COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE expense_date >= ? AND expense_date <= ? AND deleted_at IS NULL
		GROUP BY category_id
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals by category: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]float64)
	for rows.Next() {
		var categoryID int64
		var total float64
		if err := rows.Scan(&categoryID, &total); err != nil {
			return nil, fmt.Errorf("failed to scan category total: %w", err)
		}
		result[categoryID] = total
	}

	return result, nil
}

// GetTotalByContractor отримує загальну суму витрат по контрагентах.
func (r *ExpenseRepository) GetTotalByContractor(ctx context.Context, startDate, endDate int64) (map[int64]float64, error) {
	query := `
		SELECT contractor_id, COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE expense_date >= ? AND expense_date <= ? 
		  AND contractor_id IS NOT NULL 
		  AND deleted_at IS NULL
		GROUP BY contractor_id
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals by contractor: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]float64)
	for rows.Next() {
		var contractorID int64
		var total float64
		if err := rows.Scan(&contractorID, &total); err != nil {
			return nil, fmt.Errorf("failed to scan contractor total: %w", err)
		}
		result[contractorID] = total
	}

	return result, nil
}

// GetTotalByMonth отримує загальну суму витрат по місяцях за вказаний рік.
func (r *ExpenseRepository) GetTotalByMonth(ctx context.Context, year int) (map[int]float64, error) {
	// Використовуємо SQLite datetime функції для групування по місяцях
	// Оскільки expense_date зберігається як UNIX timestamp, конвертуємо його
	query := `
		SELECT 
			CAST(strftime('%m', datetime(expense_date, 'unixepoch')) AS INTEGER) as month,
			COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE 
			strftime('%Y', datetime(expense_date, 'unixepoch')) = ?
			AND deleted_at IS NULL
		GROUP BY month
	`

	// Конвертуємо рік в рядок, бо strftime повертає рядок
	yearStr := fmt.Sprintf("%d", year)

	rows, err := r.db.QueryContext(ctx, query, yearStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals by month: %w", err)
	}
	defer rows.Close()

	result := make(map[int]float64)
	for rows.Next() {
		var month int
		var total float64
		if err := rows.Scan(&month, &total); err != nil {
			return nil, fmt.Errorf("failed to scan month total: %w", err)
		}
		result[month] = total
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Приватні допоміжні методи
// ----------------------------------------------------------------------------

// scanExpenses сканує набір рядків у зріз витрат.
func (r *ExpenseRepository) scanExpenses(rows *sql.Rows) ([]*entity.Expense, error) {
	expenses := make([]*entity.Expense, 0)
	for rows.Next() {
		expense, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return expenses, nil
}

// scanExpense сканує один рядок (для GetByID).
func (r *ExpenseRepository) scanExpense(ctx context.Context, query string, args ...interface{}) (*entity.Expense, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanRow(row)
}

// Scanner інтерфейс, який реалізують sql.Row та sql.Rows.
type Scanner interface {
	Scan(dest ...interface{}) error
}

// scanRow виконує безпосереднє сканування полів у структуру.
func (r *ExpenseRepository) scanRow(scanner Scanner) (*entity.Expense, error) {
	var (
		id             int64
		categoryID     int64
		contractorID   sql.NullInt64
		expenseDate    int64
		amount         float64
		description    string
		documentType   sql.NullString
		documentNumber sql.NullString
		documentDate   sql.NullInt64
		paymentStatus  string
		paidAmount     float64
		paymentDate    sql.NullInt64
		notes          sql.NullString
		approvedBy     sql.NullInt64
		approvedAt     sql.NullInt64
		deletedAt      sql.NullInt64
		createdAt      int64
		updatedAt      int64
	)

	err := scanner.Scan(
		&id, &categoryID, &contractorID, &expenseDate, &amount, &description,
		&documentType, &documentNumber, &documentDate,
		&paymentStatus, &paidAmount, &paymentDate,
		&notes, &approvedBy, &approvedAt,
		&deletedAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan expense row: %w", err)
	}

	// Мапінг з SQL типів у Domain Entity
	expense := &entity.Expense{
		ID:            id,
		CategoryID:    categoryID,
		ExpenseDate:   time.Unix(expenseDate, 0),
		Amount:        amount,
		Description:   description,
		PaymentStatus: entity.PaymentStatus(paymentStatus),
		PaidAmount:    paidAmount,
		CreatedAt:     time.Unix(createdAt, 0),
		UpdatedAt:     time.Unix(updatedAt, 0),
	}

	if contractorID.Valid {
		val := contractorID.Int64
		expense.ContractorID = &val
	}

	if documentType.Valid {
		val := documentType.String
		expense.DocumentType = &val
	}

	if documentNumber.Valid {
		val := documentNumber.String
		expense.DocumentNumber = &val
	}

	if documentDate.Valid {
		val := time.Unix(documentDate.Int64, 0)
		expense.DocumentDate = &val
	}

	if paymentDate.Valid {
		val := time.Unix(paymentDate.Int64, 0)
		expense.PaymentDate = &val
	}

	if notes.Valid {
		val := notes.String
		expense.Notes = &val
	}

	if approvedBy.Valid {
		val := approvedBy.Int64
		expense.ApprovedBy = &val
	}

	if approvedAt.Valid {
		val := time.Unix(approvedAt.Int64, 0)
		expense.ApprovedAt = &val
	}

	if deletedAt.Valid {
		val := time.Unix(deletedAt.Int64, 0)
		expense.DeletedAt = &val
	}

	return expense, nil
}
