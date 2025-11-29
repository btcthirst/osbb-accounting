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

// ContractorPaymentRepository implements repository.ContractorPaymentRepository for SQLite.
type ContractorPaymentRepository struct {
	db *sql.DB
}

// NewContractorPaymentRepository creates a new instance of ContractorPaymentRepository.
func NewContractorPaymentRepository(db *sql.DB) *ContractorPaymentRepository {
	return &ContractorPaymentRepository{db: db}
}

// Create creates a new contractor payment.
func (r *ContractorPaymentRepository) Create(ctx context.Context, payment *entity.ContractorPayment) error {
	query := `
		INSERT INTO contractor_payments (
			contractor_id, payment_date, amount, payment_method, purpose,
			period_month, period_year, receipt_number, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = now
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = now
	}

	res, err := r.db.ExecContext(ctx, query,
		payment.ContractorID,
		payment.PaymentDate.Unix(),
		payment.Amount,
		payment.PaymentMethod,
		payment.Purpose,
		payment.PeriodMonth,
		payment.PeriodYear,
		payment.ReceiptNumber,
		payment.Notes,
		payment.CreatedAt.Unix(),
		payment.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create contractor payment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	payment.ID = id

	return nil
}

// GetByID gets a contractor payment by ID.
func (r *ContractorPaymentRepository) GetByID(ctx context.Context, id int64) (*entity.ContractorPayment, error) {
	query := `
		SELECT id, contractor_id, payment_date, amount, payment_method, purpose,
		       period_month, period_year, receipt_number, notes,
		       deleted_at, created_at, updated_at
		FROM contractor_payments
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanContractorPayment(row)
}

// Update updates a contractor payment.
func (r *ContractorPaymentRepository) Update(ctx context.Context, payment *entity.ContractorPayment) error {
	query := `
		UPDATE contractor_payments
		SET contractor_id = ?, payment_date = ?, amount = ?, payment_method = ?, purpose = ?,
		    period_month = ?, period_year = ?, receipt_number = ?, notes = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	payment.UpdatedAt = time.Now()

	res, err := r.db.ExecContext(ctx, query,
		payment.ContractorID,
		payment.PaymentDate.Unix(),
		payment.Amount,
		payment.PaymentMethod,
		payment.Purpose,
		payment.PeriodMonth,
		payment.PeriodYear,
		payment.ReceiptNumber,
		payment.Notes,
		payment.UpdatedAt.Unix(),
		payment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update contractor payment: %w", err)
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

// SoftDelete soft deletes a contractor payment.
func (r *ContractorPaymentRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE contractor_payments SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete contractor payment: %w", err)
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

// Restore restores a soft deleted contractor payment.
func (r *ContractorPaymentRepository) Restore(ctx context.Context, id int64) error {
	query := `UPDATE contractor_payments SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore contractor payment: %w", err)
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

// List gets a list of contractor payments with filtering.
func (r *ContractorPaymentRepository) List(ctx context.Context, filter repository.ContractorPaymentFilter) ([]*entity.ContractorPayment, error) {
	query, args := buildContractorPaymentListQuery(filter, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list contractor payments: %w", err)
	}
	defer rows.Close()

	var payments []*entity.ContractorPayment
	for rows.Next() {
		payment, err := scanContractorPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contractor payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// Count gets the total count of contractor payments.
func (r *ContractorPaymentRepository) Count(ctx context.Context, filter repository.ContractorPaymentFilter) (int64, error) {
	query, args := buildContractorPaymentListQuery(filter, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count contractor payments: %w", err)
	}

	return count, nil
}

// GetByContractorID gets all payments for a contractor.
func (r *ContractorPaymentRepository) GetByContractorID(ctx context.Context, contractorID int64) ([]*entity.ContractorPayment, error) {
	return r.List(ctx, repository.ContractorPaymentFilter{ContractorID: &contractorID})
}

// GetTotalByPeriod gets the total amount of contractor payments for a period.
func (r *ContractorPaymentRepository) GetTotalByPeriod(ctx context.Context, month, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) FROM contractor_payments
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

func scanContractorPayment(scanner interface {
	Scan(dest ...interface{}) error
}) (*entity.ContractorPayment, error) {
	var payment entity.ContractorPayment
	var deletedAt sql.NullInt64
	var createdAt, updatedAt, paymentDate int64

	err := scanner.Scan(
		&payment.ID,
		&payment.ContractorID,
		&paymentDate,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Purpose,
		&payment.PeriodMonth,
		&payment.PeriodYear,
		&payment.ReceiptNumber,
		&payment.Notes,
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

	return &payment, nil
}

func buildContractorPaymentListQuery(filter repository.ContractorPaymentFilter, countOnly bool) (string, []interface{}) {
	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM contractor_payments cp"
	} else {
		query = `
			SELECT cp.id, cp.contractor_id, cp.payment_date, cp.amount, cp.payment_method, cp.purpose,
			       cp.period_month, cp.period_year, cp.receipt_number, cp.notes,
			       cp.deleted_at, cp.created_at, cp.updated_at
			FROM contractor_payments cp
		`
	}

	var whereClauses []string
	var args []interface{}

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "cp.deleted_at IS NULL")
	}
	if filter.ContractorID != nil {
		whereClauses = append(whereClauses, "cp.contractor_id = ?")
		args = append(args, *filter.ContractorID)
	}
	if filter.PaymentMethod != nil {
		whereClauses = append(whereClauses, "cp.payment_method = ?")
		args = append(args, *filter.PaymentMethod)
	}
	if filter.PeriodMonth != nil {
		whereClauses = append(whereClauses, "cp.period_month = ?")
		args = append(args, *filter.PeriodMonth)
	}
	if filter.PeriodYear != nil {
		whereClauses = append(whereClauses, "cp.period_year = ?")
		args = append(args, *filter.PeriodYear)
	}
	if filter.StartDate != nil {
		whereClauses = append(whereClauses, "cp.payment_date >= ?")
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		whereClauses = append(whereClauses, "cp.payment_date <= ?")
		args = append(args, *filter.EndDate)
	}
	if filter.MinAmount != nil {
		whereClauses = append(whereClauses, "cp.amount >= ?")
		args = append(args, *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		whereClauses = append(whereClauses, "cp.amount <= ?")
		args = append(args, *filter.MaxAmount)
	}
	if filter.SearchQuery != "" {
		whereClauses = append(whereClauses, "(cp.purpose LIKE ? OR cp.receipt_number LIKE ?)")
		args = append(args, "%"+filter.SearchQuery+"%", "%"+filter.SearchQuery+"%")
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
			col := "cp.created_at"
			switch filter.OrderBy {
			case "payment_date":
				col = "cp.payment_date"
			case "amount":
				col = "cp.amount"
			case "period":
				col = "cp.period_year " + orderDir + ", cp.period_month"
			}
			if filter.OrderBy == "period" {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			} else {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			}
		} else {
			query += " ORDER BY cp.payment_date DESC"
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
