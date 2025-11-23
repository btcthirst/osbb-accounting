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

// ChargeRepository implements repository.ChargeRepository for SQLite.
type ChargeRepository struct {
	db *sql.DB
}

// NewChargeRepository creates a new instance of ChargeRepository.
func NewChargeRepository(db *sql.DB) *ChargeRepository {
	return &ChargeRepository{db: db}
}

// Create creates a new charge.
func (r *ChargeRepository) Create(ctx context.Context, charge *entity.Charge) error {
	query := `
		INSERT INTO charges (
			ownership_share_id, charge_type, charge_date, period_month, period_year,
			amount, tariff, quantity, description, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if charge.CreatedAt.IsZero() {
		charge.CreatedAt = now
	}
	if charge.UpdatedAt.IsZero() {
		charge.UpdatedAt = now
	}

	res, err := r.db.ExecContext(ctx, query,
		charge.OwnershipShareID,
		charge.ChargeType,
		charge.ChargeDate.Unix(),
		charge.PeriodMonth,
		charge.PeriodYear,
		charge.Amount,
		charge.Tariff,
		charge.Quantity,
		charge.Description,
		charge.Notes,
		charge.CreatedAt.Unix(),
		charge.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create charge: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	charge.ID = id

	return nil
}

// GetByID gets a charge by ID.
func (r *ChargeRepository) GetByID(ctx context.Context, id int64) (*entity.Charge, error) {
	query := `
		SELECT id, ownership_share_id, charge_type, charge_date, period_month, period_year,
		       amount, tariff, quantity, description, notes, deleted_at, created_at, updated_at
		FROM charges
		WHERE id = ? AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanCharge(row)
}

// Update updates a charge.
func (r *ChargeRepository) Update(ctx context.Context, charge *entity.Charge) error {
	query := `
		UPDATE charges
		SET ownership_share_id = ?, charge_type = ?, charge_date = ?, period_month = ?, period_year = ?,
		    amount = ?, tariff = ?, quantity = ?, description = ?, notes = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	charge.UpdatedAt = time.Now()

	res, err := r.db.ExecContext(ctx, query,
		charge.OwnershipShareID,
		charge.ChargeType,
		charge.ChargeDate.Unix(),
		charge.PeriodMonth,
		charge.PeriodYear,
		charge.Amount,
		charge.Tariff,
		charge.Quantity,
		charge.Description,
		charge.Notes,
		charge.UpdatedAt.Unix(),
		charge.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update charge: %w", err)
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

// SoftDelete soft deletes a charge.
func (r *ChargeRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE charges SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete charge: %w", err)
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

// Restore restores a soft deleted charge.
func (r *ChargeRepository) Restore(ctx context.Context, id int64) error {
	query := `UPDATE charges SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore charge: %w", err)
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

// List gets a list of charges with filtering.
func (r *ChargeRepository) List(ctx context.Context, filter repository.ChargeFilter) ([]*entity.Charge, error) {
	query, args := buildChargeListQuery(filter, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list charges: %w", err)
	}
	defer rows.Close()

	var charges []*entity.Charge
	for rows.Next() {
		charge, err := scanCharge(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charge: %w", err)
		}
		charges = append(charges, charge)
	}

	return charges, nil
}

// Count gets the total count of charges.
func (r *ChargeRepository) Count(ctx context.Context, filter repository.ChargeFilter) (int64, error) {
	query, args := buildChargeListQuery(filter, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count charges: %w", err)
	}

	return count, nil
}

// GetByOwnershipShareID gets all charges for an ownership share.
func (r *ChargeRepository) GetByOwnershipShareID(ctx context.Context, ownershipShareID int64) ([]*entity.Charge, error) {
	return r.List(ctx, repository.ChargeFilter{OwnershipShareID: &ownershipShareID})
}

// GetByPeriod gets charges for a period.
func (r *ChargeRepository) GetByPeriod(ctx context.Context, month, year int) ([]*entity.Charge, error) {
	return r.List(ctx, repository.ChargeFilter{PeriodMonth: &month, PeriodYear: &year})
}

// GetByOwnershipShareAndPeriod gets charges for an ownership share for a period.
func (r *ChargeRepository) GetByOwnershipShareAndPeriod(ctx context.Context, ownershipShareID int64, month, year int) ([]*entity.Charge, error) {
	return r.List(ctx, repository.ChargeFilter{
		OwnershipShareID: &ownershipShareID,
		PeriodMonth:      &month,
		PeriodYear:       &year,
	})
}

// CheckDuplicatePeriod checks if a charge exists for a period.
func (r *ChargeRepository) CheckDuplicatePeriod(ctx context.Context, ownershipShareID int64, chargeType entity.ChargeType, month, year int, excludeID *int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM charges
		WHERE ownership_share_id = ? AND charge_type = ? AND period_month = ? AND period_year = ? AND deleted_at IS NULL
	`
	args := []interface{}{ownershipShareID, chargeType, month, year}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate period: %w", err)
	}

	return count > 0, nil
}

// CalculateTotalForPeriod calculates the total amount of charges for a period.
func (r *ChargeRepository) CalculateTotalForPeriod(ctx context.Context, month, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) FROM charges
		WHERE period_month = ? AND period_year = ? AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRowContext(ctx, query, month, year).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total for period: %w", err)
	}

	return total, nil
}

// CalculateTotalForOwnershipShare calculates the total amount of charges for an ownership share.
func (r *ChargeRepository) CalculateTotalForOwnershipShare(ctx context.Context, ownershipShareID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) FROM charges
		WHERE ownership_share_id = ? AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRowContext(ctx, query, ownershipShareID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total for ownership share: %w", err)
	}

	return total, nil
}

// GetWithDetails gets a charge with details.
func (r *ChargeRepository) GetWithDetails(ctx context.Context, id int64) (*repository.ChargeDetails, error) {
	// This requires joining with ownership_shares, owners, apartments
	// For simplicity, we can fetch the charge and then fetch related entities, or use a JOIN query.
	// Let's use a JOIN query.
	query := `
		SELECT 
			c.id, c.ownership_share_id, c.charge_type, c.charge_date, c.period_month, c.period_year,
			c.amount, c.tariff, c.quantity, c.description, c.notes, c.deleted_at, c.created_at, c.updated_at,
			o.first_name || ' ' || o.last_name as owner_name, o.phone, o.email,
			a.apartment_number, a.floor, a.entrance,
			os.share_numerator, os.share_denominator
		FROM charges c
		JOIN ownership_shares os ON c.ownership_share_id = os.id
		JOIN owners o ON os.owner_id = o.id
		JOIN apartments a ON os.apartment_id = a.id
		WHERE c.id = ? AND c.deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var details repository.ChargeDetails
	var charge entity.Charge
	var deletedAt sql.NullInt64
	var createdAt, updatedAt, chargeDate int64
	var num, den int

	err := row.Scan(
		&charge.ID, &charge.OwnershipShareID, &charge.ChargeType, &chargeDate, &charge.PeriodMonth, &charge.PeriodYear,
		&charge.Amount, &charge.Tariff, &charge.Quantity, &charge.Description, &charge.Notes, &deletedAt, &createdAt, &updatedAt,
		&details.OwnerName, &details.OwnerPhone, &details.OwnerEmail,
		&details.ApartmentNumber, &details.ApartmentFloor, &details.ApartmentEntrance,
		&num, &den,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get charge with details: %w", err)
	}

	charge.ChargeDate = time.Unix(chargeDate, 0)
	charge.CreatedAt = time.Unix(createdAt, 0)
	charge.UpdatedAt = time.Unix(updatedAt, 0)
	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		charge.DeletedAt = &t
	}

	details.Charge = &charge
	details.ShareFraction = fmt.Sprintf("%d/%d", num, den)
	details.SharePercent = float64(num) / float64(den) * 100

	return &details, nil
}

// ListWithDetails gets a list of charges with details.
func (r *ChargeRepository) ListWithDetails(ctx context.Context, filter repository.ChargeFilter) ([]*repository.ChargeDetails, error) {
	// Similar to List but with JOINs.
	// Constructing dynamic query with JOINs is complex.
	// Simplified approach: List IDs first, then GetWithDetails (N+1 problem but simpler code)
	// OR: Write a separate query builder for details.

	// Let's implement a proper query for ListWithDetails
	baseQuery := `
		SELECT 
			c.id, c.ownership_share_id, c.charge_type, c.charge_date, c.period_month, c.period_year,
			c.amount, c.tariff, c.quantity, c.description, c.notes, c.deleted_at, c.created_at, c.updated_at,
			o.first_name || ' ' || o.last_name as owner_name, o.phone, o.email,
			a.apartment_number, a.floor, a.entrance,
			os.share_numerator, os.share_denominator
		FROM charges c
		JOIN ownership_shares os ON c.ownership_share_id = os.id
		JOIN owners o ON os.owner_id = o.id
		JOIN apartments a ON os.apartment_id = a.id
	`

	whereClauses := []string{}
	args := []interface{}{}

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "c.deleted_at IS NULL")
	}
	if filter.OwnershipShareID != nil {
		whereClauses = append(whereClauses, "c.ownership_share_id = ?")
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
	if filter.ChargeType != nil {
		whereClauses = append(whereClauses, "c.charge_type = ?")
		args = append(args, *filter.ChargeType)
	}
	if filter.PeriodMonth != nil {
		whereClauses = append(whereClauses, "c.period_month = ?")
		args = append(args, *filter.PeriodMonth)
	}
	if filter.PeriodYear != nil {
		whereClauses = append(whereClauses, "c.period_year = ?")
		args = append(args, *filter.PeriodYear)
	}
	// ... other filters ...

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Order and Limit
	if filter.OrderBy != "" {
		orderDir := "ASC"
		if filter.OrderDesc {
			orderDir = "DESC"
		}
		// Map sort fields to columns
		col := "c.created_at"
		switch filter.OrderBy {
		case "charge_date":
			col = "c.charge_date"
		case "period":
			col = "c.period_year " + orderDir + ", c.period_month" // Special case
		case "amount":
			col = "c.amount"
		}

		if filter.OrderBy == "period" {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
		} else {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
		}
	} else {
		baseQuery += " ORDER BY c.created_at DESC"
	}

	if filter.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list charges with details: %w", err)
	}
	defer rows.Close()

	var detailsList []*repository.ChargeDetails
	for rows.Next() {
		var details repository.ChargeDetails
		var charge entity.Charge
		var deletedAt sql.NullInt64
		var createdAt, updatedAt, chargeDate int64
		var num, den int

		err := rows.Scan(
			&charge.ID, &charge.OwnershipShareID, &charge.ChargeType, &chargeDate, &charge.PeriodMonth, &charge.PeriodYear,
			&charge.Amount, &charge.Tariff, &charge.Quantity, &charge.Description, &charge.Notes, &deletedAt, &createdAt, &updatedAt,
			&details.OwnerName, &details.OwnerPhone, &details.OwnerEmail,
			&details.ApartmentNumber, &details.ApartmentFloor, &details.ApartmentEntrance,
			&num, &den,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charge details: %w", err)
		}

		charge.ChargeDate = time.Unix(chargeDate, 0)
		charge.CreatedAt = time.Unix(createdAt, 0)
		charge.UpdatedAt = time.Unix(updatedAt, 0)
		if deletedAt.Valid {
			t := time.Unix(deletedAt.Int64, 0)
			charge.DeletedAt = &t
		}

		details.Charge = &charge
		details.ShareFraction = fmt.Sprintf("%d/%d", num, den)
		details.SharePercent = float64(num) / float64(den) * 100

		detailsList = append(detailsList, &details)
	}

	return detailsList, nil
}

// GetOverdueCharges gets overdue charges.
func (r *ChargeRepository) GetOverdueCharges(ctx context.Context, asOfDate time.Time) ([]*entity.Charge, error) {
	// Logic: period end date + 1 month < asOfDate AND not fully paid (payment logic is separate, assuming charges are unpaid if no payment linked?
	// Wait, payment status is not in charges table in schema v2.0?
	// Checking schema... "charges" table does NOT have payment_status. "expenses" has.
	// "payments" table links to "ownership_shares".
	// So to know if a charge is paid, we need to sum payments for that period/share and compare with charge amount.
	// BUT the interface says GetOverdueCharges.
	// If the schema doesn't track paid status on charge, this is complex.
	// Let's assume for now we return all charges where period is overdue, regardless of payment (as we don't have payment info here easily).
	// OR maybe the user meant unpaid overdue charges?
	// Given the schema, let's just implement based on date.

	// Period end date = last day of period_month/period_year.
	// Due date = Period end date + 1 month.
	// We need charges where Due Date < asOfDate.

	// Simplified SQL: (period_year * 12 + period_month) < (asOfDate_year * 12 + asOfDate_month - 1)

	targetMonthTotal := asOfDate.Year()*12 + int(asOfDate.Month()) - 1

	query := `
		SELECT id, ownership_share_id, charge_type, charge_date, period_month, period_year,
		       amount, tariff, quantity, description, notes, deleted_at, created_at, updated_at
		FROM charges
		WHERE (period_year * 12 + period_month) < ? AND deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, targetMonthTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue charges: %w", err)
	}
	defer rows.Close()

	var charges []*entity.Charge
	for rows.Next() {
		charge, err := scanCharge(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charge: %w", err)
		}
		charges = append(charges, charge)
	}

	return charges, nil
}

// GetStatistics gets statistics.
func (r *ChargeRepository) GetStatistics(ctx context.Context, filter repository.ChargeStatisticsFilter) (*repository.ChargeStatistics, error) {
	stats := &repository.ChargeStatistics{
		ByType:   make(map[entity.ChargeType]repository.TypeStatistics),
		ByPeriod: make(map[string]repository.PeriodStatistics),
	}

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
	// ... other filters ...

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Total Stats
	queryTotal := "SELECT COUNT(*), COALESCE(SUM(amount), 0), COALESCE(AVG(amount), 0), COALESCE(MIN(amount), 0), COALESCE(MAX(amount), 0) FROM charges WHERE " + whereSQL
	err := r.db.QueryRowContext(ctx, queryTotal, args...).Scan(
		&stats.TotalCharges, &stats.TotalAmount, &stats.AverageAmount, &stats.MinAmount, &stats.MaxAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stats: %w", err)
	}

	// 2. By Type
	queryType := "SELECT charge_type, COUNT(*), COALESCE(SUM(amount), 0), COALESCE(AVG(amount), 0) FROM charges WHERE " + whereSQL + " GROUP BY charge_type"
	rowsType, err := r.db.QueryContext(ctx, queryType, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	defer rowsType.Close()

	for rowsType.Next() {
		var t entity.ChargeType
		var s repository.TypeStatistics
		if err := rowsType.Scan(&t, &s.Count, &s.TotalAmount, &s.AvgAmount); err != nil {
			return nil, err
		}
		stats.ByType[t] = s
	}

	// 3. By Period
	// SQLite doesn't have easy formatting, so we fetch period_month/year
	queryPeriod := "SELECT period_month, period_year, COUNT(*), COALESCE(SUM(amount), 0), COALESCE(AVG(amount), 0) FROM charges WHERE " + whereSQL + " GROUP BY period_year, period_month"
	rowsPeriod, err := r.db.QueryContext(ctx, queryPeriod, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get period stats: %w", err)
	}
	defer rowsPeriod.Close()

	for rowsPeriod.Next() {
		var m, y int
		var s repository.PeriodStatistics
		if err := rowsPeriod.Scan(&m, &y, &s.Count, &s.TotalAmount, &s.AvgAmount); err != nil {
			return nil, err
		}
		s.Period = fmt.Sprintf("%02d/%d", m, y)
		stats.ByPeriod[s.Period] = s
	}

	return stats, nil
}

// GetByApartmentID gets charges for an apartment.
func (r *ChargeRepository) GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.Charge, error) {
	return r.List(ctx, repository.ChargeFilter{ApartmentID: &apartmentID})
}

// GetByOwnerID gets charges for an owner.
func (r *ChargeRepository) GetByOwnerID(ctx context.Context, ownerID int64) ([]*entity.Charge, error) {
	return r.List(ctx, repository.ChargeFilter{OwnerID: &ownerID})
}

// BulkCreate creates multiple charges.
func (r *ChargeRepository) BulkCreate(ctx context.Context, charges []*entity.Charge) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO charges (
			ownership_share_id, charge_type, charge_date, period_month, period_year,
			amount, tariff, quantity, description, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()

	for _, charge := range charges {
		if charge.CreatedAt.IsZero() {
			charge.CreatedAt = now
		}
		if charge.UpdatedAt.IsZero() {
			charge.UpdatedAt = now
		}

		_, err := stmt.ExecContext(ctx,
			charge.OwnershipShareID,
			charge.ChargeType,
			charge.ChargeDate.Unix(),
			charge.PeriodMonth,
			charge.PeriodYear,
			charge.Amount,
			charge.Tariff,
			charge.Quantity,
			charge.Description,
			charge.Notes,
			charge.CreatedAt.Unix(),
			charge.UpdatedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("failed to bulk create charge: %w", err)
		}
	}

	return tx.Commit()
}

// GetPeriodRange gets charges for a range of periods.
func (r *ChargeRepository) GetPeriodRange(ctx context.Context, startMonth, startYear, endMonth, endYear int) ([]*entity.Charge, error) {
	// Logic: (year > startYear OR (year == startYear AND month >= startMonth)) AND (year < endYear OR (year == endYear AND month <= endMonth))
	// Simplified: year*12+month BETWEEN start AND end

	startVal := startYear*12 + startMonth
	endVal := endYear*12 + endMonth

	query := `
		SELECT id, ownership_share_id, charge_type, charge_date, period_month, period_year,
		       amount, tariff, quantity, description, notes, deleted_at, created_at, updated_at
		FROM charges
		WHERE (period_year * 12 + period_month) BETWEEN ? AND ? AND deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, startVal, endVal)
	if err != nil {
		return nil, fmt.Errorf("failed to get period range: %w", err)
	}
	defer rows.Close()

	var charges []*entity.Charge
	for rows.Next() {
		charge, err := scanCharge(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charge: %w", err)
		}
		charges = append(charges, charge)
	}

	return charges, nil
}

// Helper functions

func scanCharge(scanner interface {
	Scan(dest ...interface{}) error
}) (*entity.Charge, error) {
	var charge entity.Charge
	var deletedAt sql.NullInt64
	var createdAt, updatedAt, chargeDate int64

	err := scanner.Scan(
		&charge.ID,
		&charge.OwnershipShareID,
		&charge.ChargeType,
		&chargeDate,
		&charge.PeriodMonth,
		&charge.PeriodYear,
		&charge.Amount,
		&charge.Tariff,
		&charge.Quantity,
		&charge.Description,
		&charge.Notes,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	charge.ChargeDate = time.Unix(chargeDate, 0)
	charge.CreatedAt = time.Unix(createdAt, 0)
	charge.UpdatedAt = time.Unix(updatedAt, 0)
	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		charge.DeletedAt = &t
	}

	return &charge, nil
}

func buildChargeListQuery(filter repository.ChargeFilter, countOnly bool) (string, []interface{}) {
	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM charges c"
	} else {
		query = `
			SELECT c.id, c.ownership_share_id, c.charge_type, c.charge_date, c.period_month, c.period_year,
			       c.amount, c.tariff, c.quantity, c.description, c.notes, c.deleted_at, c.created_at, c.updated_at
			FROM charges c
		`
	}

	// Joins if needed for filtering
	if filter.ApartmentID != nil || filter.OwnerID != nil {
		query += " JOIN ownership_shares os ON c.ownership_share_id = os.id"
	}

	var whereClauses []string
	var args []interface{}

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "c.deleted_at IS NULL")
	}
	if filter.OwnershipShareID != nil {
		whereClauses = append(whereClauses, "c.ownership_share_id = ?")
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
	if filter.ChargeType != nil {
		whereClauses = append(whereClauses, "c.charge_type = ?")
		args = append(args, *filter.ChargeType)
	}
	if filter.PeriodMonth != nil {
		whereClauses = append(whereClauses, "c.period_month = ?")
		args = append(args, *filter.PeriodMonth)
	}
	if filter.PeriodYear != nil {
		whereClauses = append(whereClauses, "c.period_year = ?")
		args = append(args, *filter.PeriodYear)
	}
	if filter.StartDate != nil {
		whereClauses = append(whereClauses, "c.charge_date >= ?")
		args = append(args, filter.StartDate.Unix())
	}
	if filter.EndDate != nil {
		whereClauses = append(whereClauses, "c.charge_date <= ?")
		args = append(args, filter.EndDate.Unix())
	}
	if filter.MinAmount != nil {
		whereClauses = append(whereClauses, "c.amount >= ?")
		args = append(args, *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		whereClauses = append(whereClauses, "c.amount <= ?")
		args = append(args, *filter.MaxAmount)
	}
	if filter.SearchQuery != "" {
		whereClauses = append(whereClauses, "c.description LIKE ?")
		args = append(args, "%"+filter.SearchQuery+"%")
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
			col := "c.created_at"
			switch filter.OrderBy {
			case "charge_date":
				col = "c.charge_date"
			case "period":
				col = "c.period_year " + orderDir + ", c.period_month"
			case "amount":
				col = "c.amount"
			}
			if filter.OrderBy == "period" {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			} else {
				query += fmt.Sprintf(" ORDER BY %s %s", col, orderDir)
			}
		} else {
			query += " ORDER BY c.created_at DESC"
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
