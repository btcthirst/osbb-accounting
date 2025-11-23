package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

// PaymentRepository implements repository.PaymentRepository for SQLite.
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new instance of PaymentRepository.
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment.
func (r *PaymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	query := `
		INSERT INTO payments (
			ownership_share_id, payment_date, amount, payment_method, payment_purpose,
			period_month, period_year, receipt_number, notes, approved_by, approved_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = now
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = now
	}

	var approvedAtUnix *int64
	if payment.ApprovedAt != nil {
		t := payment.ApprovedAt.Unix()
		approvedAtUnix = &t
	}

	res, err := r.db.ExecContext(ctx, query,
		payment.OwnershipShareID,
		payment.PaymentDate.Unix(),
		payment.Amount,
		payment.PaymentMethod,
		payment.PaymentPurpose,
		payment.PeriodMonth,
		payment.PeriodYear,
		payment.ReceiptNumber,
		payment.Notes,
		payment.ApprovedBy,
		approvedAtUnix,
		payment.CreatedAt.Unix(),
		payment.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	payment.ID = id

	return nil
}

// GetByID gets a payment by ID.
func (r *PaymentRepository) GetByID(ctx context.Context, id int64) (*entity.Payment, error) {
	query := `
		SELECT id, ownership_share_id, payment_date, amount, payment_method, payment_purpose,
		       period_month, period_year, receipt_number, notes, approved_by, approved_at,
		       deleted_at, created_at, updated_at
		FROM payments
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanPayment(row)
}

// Update updates a payment.
func (r *PaymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	query := `
		UPDATE payments
		SET ownership_share_id = ?, payment_date = ?, amount = ?, payment_method = ?, payment_purpose = ?,
		    period_month = ?, period_year = ?, receipt_number = ?, notes = ?, approved_by = ?, approved_at = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	payment.UpdatedAt = time.Now()

	var approvedAtUnix *int64
	if payment.ApprovedAt != nil {
		t := payment.ApprovedAt.Unix()
		approvedAtUnix = &t
	}

	res, err := r.db.ExecContext(ctx, query,
		payment.OwnershipShareID,
		payment.PaymentDate.Unix(),
		payment.Amount,
		payment.PaymentMethod,
		payment.PaymentPurpose,
		payment.PeriodMonth,
		payment.PeriodYear,
		payment.ReceiptNumber,
		payment.Notes,
		payment.ApprovedBy,
		approvedAtUnix,
		payment.UpdatedAt.Unix(),
		payment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SoftDelete soft deletes a payment.
func (r *PaymentRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE payments SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete payment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Restore restores a soft deleted payment.
func (r *PaymentRepository) Restore(ctx context.Context, id int64) error {
	query := `UPDATE payments SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore payment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List gets a list of payments with filtering.
func (r *PaymentRepository) List(ctx context.Context, filter repository.PaymentFilter) ([]*entity.Payment, error) {
	query, args := buildPaymentListQuery(filter, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*entity.Payment
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// Count gets the total count of payments.
func (r *PaymentRepository) Count(ctx context.Context, filter repository.PaymentFilter) (int64, error) {
	query, args := buildPaymentListQuery(filter, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count payments: %w", err)
	}

	return count, nil
}

// GetByOwnershipShareID gets all payments for an ownership share.
func (r *PaymentRepository) GetByOwnershipShareID(ctx context.Context, shareID int64) ([]*entity.Payment, error) {
	return r.List(ctx, repository.PaymentFilter{OwnershipShareID: &shareID})
}

// GetStatistics gets aggregated statistics for payments.
func (r *PaymentRepository) GetStatistics(ctx context.Context, filter repository.PaymentFilter) (*repository.PaymentStatistics, error) {
	stats := &repository.PaymentStatistics{}

	// Base WHERE clause
	whereClauses := []string{"deleted_at IS NULL"}
	args := []interface{}{}

	if filter.PeriodMonth != nil {
		whereClauses = append(whereClauses, "period_month = ?")
		args = append(args, *filter.PeriodMonth)
	}
	if filter.PeriodYear != nil {
		whereClauses = append(whereClauses, "period_year = ?")
		args = append(args, *filter.PeriodYear)
	}
	// ... other filters if needed for stats ...

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Total Stats
	queryTotal := `
		SELECT 
			COUNT(*), 
			COALESCE(SUM(amount), 0), 
			COALESCE(AVG(amount), 0), 
			COALESCE(MIN(amount), 0), 
			COALESCE(MAX(amount), 0) 
		FROM payments WHERE ` + whereSQL

	err := r.db.QueryRowContext(ctx, queryTotal, args...).Scan(
		&stats.TotalCount, &stats.TotalAmount, &stats.AverageAmount, &stats.MinAmount, &stats.MaxAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stats: %w", err)
	}

	// 2. By Method (Count)
	queryMethodCount := `
		SELECT payment_method, COUNT(*) 
		FROM payments 
		WHERE ` + whereSQL + ` 
		GROUP BY payment_method
	`
	rowsCount, err := r.db.QueryContext(ctx, queryMethodCount, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get method counts: %w", err)
	}
	defer rowsCount.Close()

	for rowsCount.Next() {
		var method entity.PaymentMethod
		var count int64
		if err := rowsCount.Scan(&method, &count); err != nil {
			return nil, err
		}
		switch method {
		case entity.PaymentMethodCash:
			stats.CountByCash = count
		case entity.PaymentMethodCard:
			stats.CountByCard = count
		case entity.PaymentMethodBankTransfer:
			stats.CountByBankTransfer = count
		case entity.PaymentMethodOther:
			stats.CountByOther = count
		}
	}

	// 3. By Method (Amount)
	queryMethodAmount := `
		SELECT payment_method, COALESCE(SUM(amount), 0)
		FROM payments 
		WHERE ` + whereSQL + ` 
		GROUP BY payment_method
	`
	rowsAmount, err := r.db.QueryContext(ctx, queryMethodAmount, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get method amounts: %w", err)
	}
	defer rowsAmount.Close()

	for rowsAmount.Next() {
		var method entity.PaymentMethod
		var amount float64
		if err := rowsAmount.Scan(&method, &amount); err != nil {
			return nil, err
		}
		switch method {
		case entity.PaymentMethodCash:
			stats.AmountByCash = amount
		case entity.PaymentMethodCard:
			stats.AmountByCard = amount
		case entity.PaymentMethodBankTransfer:
			stats.AmountByBankTransfer = amount
		case entity.PaymentMethodOther:
			stats.AmountByOther = amount
		}
	}

	// 4. Approved/Not Approved
	queryApproved := `
		SELECT 
			COUNT(CASE WHEN approved_by IS NOT NULL THEN 1 END),
			COUNT(CASE WHEN approved_by IS NULL THEN 1 END)
		FROM payments 
		WHERE ` + whereSQL

	err = r.db.QueryRowContext(ctx, queryApproved, args...).Scan(&stats.ApprovedCount, &stats.NotApprovedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved stats: %w", err)
	}

	return stats, nil
}

// GetTotalByPeriod gets the total amount of payments for a period.
func (r *PaymentRepository) GetTotalByPeriod(ctx context.Context, month, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) FROM payments
		WHERE period_month = ? AND period_year = ? AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRowContext(ctx, query, month, year).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total by period: %w", err)
	}

	return total, nil
}

// Helper functions

func scanPayment(scanner interface {
	Scan(dest ...interface{}) error
}) (*entity.Payment, error) {
	var payment entity.Payment
	var deletedAt, approvedAt sql.NullInt64
	var createdAt, updatedAt, paymentDate int64
	var approvedBy sql.NullInt64

	err := scanner.Scan(
		&payment.ID,
		&payment.OwnershipShareID,
		&paymentDate,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.PaymentPurpose,
		&payment.PeriodMonth,
		&payment.PeriodYear,
		&payment.ReceiptNumber,
		&payment.Notes,
		&approvedBy,
		&approvedAt,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	payment.PaymentDate = time.Unix(paymentDate, 0)
	payment.CreatedAt = time.Unix(createdAt, 0)
	payment.UpdatedAt = time.Unix(updatedAt, 0)

	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		payment.DeletedAt = &t
	}
	if approvedAt.Valid {
		t := time.Unix(approvedAt.Int64, 0)
		payment.ApprovedAt = &t
	}
	if approvedBy.Valid {
		id := approvedBy.Int64
		payment.ApprovedBy = &id
	}

	return &payment, nil
}

func buildPaymentListQuery(filter repository.PaymentFilter, countOnly bool) (string, []interface{}) {
	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM payments p"
	} else {
		query = `
			SELECT p.id, p.ownership_share_id, p.payment_date, p.amount, p.payment_method, p.payment_purpose,
			       p.period_month, p.period_year, p.receipt_number, p.notes, p.approved_by, p.approved_at,
			       p.deleted_at, p.created_at, p.updated_at
			FROM payments p
		`
	}

	// Joins if needed
	if filter.ApartmentID != nil || filter.OwnerID != nil {
		query += " JOIN ownership_shares os ON p.ownership_share_id = os.id"
	}

	var whereClauses []string
	var args []interface{}

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
	}
	if filter.OwnershipShareID != nil {
		whereClauses = append(whereClauses, "p.ownership_share_id = ?")
		args = append(args, *filter.OwnershipShareID)
	}
	if filter.ApartmentID != nil {
		whereClauses = append(whereClauses, "os.apartment_id = ?")
		args = append(args, *filter.ApartmentID)
	}
	if filter.OwnerID != nil {
		whereClauses = append(whereClauses, "os.owner_id = ?")
		args = append(args, *filter.OwnerID)
	}
	if filter.PaymentMethod != nil {
		whereClauses = append(whereClauses, "p.payment_method = ?")
		args = append(args, *filter.PaymentMethod)
	}
	if filter.PeriodMonth != nil {
		whereClauses = append(whereClauses, "p.period_month = ?")
		args = append(args, *filter.PeriodMonth)
	}
	if filter.PeriodYear != nil {
		whereClauses = append(whereClauses, "p.period_year = ?")
		args = append(args, *filter.PeriodYear)
	}
	if filter.StartDate != nil {
		whereClauses = append(whereClauses, "p.payment_date >= ?")
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		whereClauses = append(whereClauses, "p.payment_date <= ?")
		args = append(args, *filter.EndDate)
	}
	if filter.MinAmount != nil {
		whereClauses = append(whereClauses, "p.amount >= ?")
		args = append(args, *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		whereClauses = append(whereClauses, "p.amount <= ?")
		args = append(args, *filter.MaxAmount)
	}
	if filter.SearchQuery != "" {
		whereClauses = append(whereClauses, "(p.payment_purpose LIKE ? OR p.receipt_number LIKE ?)")
		args = append(args, "%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%")
	}
	if filter.IsApproved != nil {
		if *filter.IsApproved {
			whereClauses = append(whereClauses, "p.approved_by IS NOT NULL")
		} else {
			whereClauses = append(whereClauses, "p.approved_by IS NULL")
		}
	}
	if filter.ApprovedBy != nil {
		whereClauses = append(whereClauses, "p.approved_by = ?")
		args = append(args, *filter.ApprovedBy)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if !countOnly {
		if filter.OrderBy != "" {
			orderDir := "ASC"
			if filter.OrderDesc {
				orderDir = "DESC"
			}
			col := "p.created_at"
			switch filter.OrderBy {
			case "payment_date":
				col = "p.payment_date"
			case "amount":
				col = "p.amount"
			case "period":
				col = "p.period_year " + orderDir + ", p.period_month"
			}
			if filter.OrderBy == "period" {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			} else {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			}
		} else {
			query += " ORDER BY p.created_at DESC"
		}

		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	return query, args
}
