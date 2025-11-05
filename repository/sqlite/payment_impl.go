package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"
)

// SQLitePaymentRepository - SQLite реалізація PaymentRepository.
type SQLitePaymentRepository struct {
	db *sql.DB
}

// NewSQLitePaymentRepository створює новий екземпляр SQLite репозиторію платежів.
func NewSQLitePaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &SQLitePaymentRepository{db: db}
}

// Save реалізує збереження нового платежу.
func (r *SQLitePaymentRepository) Save(payment *domain.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO payments (
			apartment_id, type, category, amount, description,
			payment_date, period, created_by, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		payment.ApartmentID,
		payment.Type,
		payment.Category,
		payment.Amount,
		payment.Description,
		payment.PaymentDate,
		payment.Period,
		payment.CreatedBy,
		payment.Notes,
		now,
		now,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	payment.ID = int(id)
	return nil
}

// FindByID реалізує пошук платежу за ID.
func (r *SQLitePaymentRepository) FindByID(id int) (*domain.Payment, error) {
	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		WHERE id = ?
	`

	payment := &domain.Payment{}
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.ApartmentID,
		&payment.Type,
		&payment.Category,
		&payment.Amount,
		&payment.Description,
		&payment.PaymentDate,
		&payment.Period,
		&payment.CreatedBy,
		&notes,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	if notes.Valid {
		payment.Notes = notes.String
	}

	return payment, nil
}

// FindByApartmentID реалізує пошук платежів квартири.
func (r *SQLitePaymentRepository) FindByApartmentID(apartmentID int) ([]*domain.Payment, error) {
	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		WHERE apartment_id = ?
		ORDER BY payment_date DESC, id DESC
	`

	rows, err := r.db.Query(query, apartmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// FindByFilter реалізує пошук з фільтрами.
func (r *SQLitePaymentRepository) FindByFilter(filter *domain.PaymentFilter) ([]*domain.Payment, error) {
	if filter == nil {
		filter = domain.DefaultPaymentFilter()
	}

	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.ApartmentID != nil {
		query += ` AND apartment_id = ?`
		args = append(args, *filter.ApartmentID)
	}

	if filter.Type != nil {
		query += ` AND type = ?`
		args = append(args, *filter.Type)
	}

	if filter.Category != nil {
		query += ` AND category = ?`
		args = append(args, *filter.Category)
	}

	if filter.Period != nil {
		query += ` AND period = ?`
		args = append(args, *filter.Period)
	}

	if filter.DateFrom != nil {
		query += ` AND payment_date >= ?`
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		query += ` AND payment_date <= ?`
		args = append(args, *filter.DateTo)
	}

	if filter.MinAmount != nil {
		query += ` AND amount >= ?`
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		query += ` AND amount <= ?`
		args = append(args, *filter.MaxAmount)
	}

	if filter.CreatedBy != nil {
		query += ` AND created_by = ?`
		args = append(args, *filter.CreatedBy)
	}

	// Сортування
	sortBy := "payment_date"
	if filter.SortBy != "" {
		validFields := map[string]bool{
			"payment_date": true,
			"amount":       true,
			"type":         true,
			"category":     true,
			"period":       true,
		}
		if validFields[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}

	query += fmt.Sprintf(" ORDER BY %s", sortBy)
	if filter.SortDesc {
		query += " DESC"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// FindByPeriod реалізує пошук платежів за період.
func (r *SQLitePaymentRepository) FindByPeriod(period string) ([]*domain.Payment, error) {
	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		WHERE period = ?
		ORDER BY apartment_id, payment_date DESC
	`

	rows, err := r.db.Query(query, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// FindByApartmentAndPeriod реалізує пошук платежів квартири за період.
func (r *SQLitePaymentRepository) FindByApartmentAndPeriod(apartmentID int, period string) ([]*domain.Payment, error) {
	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		WHERE apartment_id = ? AND period = ?
		ORDER BY payment_date DESC
	`

	rows, err := r.db.Query(query, apartmentID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// Update реалізує оновлення платежу.
func (r *SQLitePaymentRepository) Update(payment *domain.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE payments
		SET apartment_id = ?, type = ?, category = ?, amount = ?,
		    description = ?, payment_date = ?, period = ?,
		    notes = ?, updated_at = ?
		WHERE id = ?
	`

	payment.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		payment.ApartmentID,
		payment.Type,
		payment.Category,
		payment.Amount,
		payment.Description,
		payment.PaymentDate,
		payment.Period,
		payment.Notes,
		payment.UpdatedAt,
		payment.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

// Delete реалізує видалення платежу.
func (r *SQLitePaymentRepository) Delete(id int) error {
	query := `DELETE FROM payments WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

// GetSummaryByApartment реалізує отримання зведення по квартирі.
func (r *SQLitePaymentRepository) GetSummaryByApartment(apartmentID int) (*domain.PaymentSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) as total_charges,
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) as total_payments,
			COUNT(CASE WHEN type = 'charge' THEN 1 END) as charges_count,
			COUNT(CASE WHEN type = 'incoming' THEN 1 END) as payments_count,
			MAX(CASE WHEN type = 'incoming' THEN payment_date END) as last_payment
		FROM payments
		WHERE apartment_id = ?
	`

	summary := &domain.PaymentSummary{
		ApartmentID: apartmentID,
	}

	var lastPayment sql.NullTime

	err := r.db.QueryRow(query, apartmentID).Scan(
		&summary.TotalCharges,
		&summary.TotalPayments,
		&summary.ChargesCount,
		&summary.PaymentsCount,
		&lastPayment,
	)

	if err != nil {
		return nil, err
	}

	if lastPayment.Valid {
		summary.LastPayment = &lastPayment.Time
	}

	// Баланс = платежі - нарахування
	summary.Balance = summary.TotalPayments - summary.TotalCharges

	return summary, nil
}

// GetSummaryByPeriod реалізує отримання зведення за період.
func (r *SQLitePaymentRepository) GetSummaryByPeriod(period string) (map[int]*domain.PaymentSummary, error) {
	query := `
		SELECT 
			apartment_id,
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) as total_charges,
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) as total_payments,
			COUNT(CASE WHEN type = 'charge' THEN 1 END) as charges_count,
			COUNT(CASE WHEN type = 'incoming' THEN 1 END) as payments_count
		FROM payments
		WHERE period = ?
		GROUP BY apartment_id
	`

	rows, err := r.db.Query(query, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make(map[int]*domain.PaymentSummary)

	for rows.Next() {
		summary := &domain.PaymentSummary{}

		err := rows.Scan(
			&summary.ApartmentID,
			&summary.TotalCharges,
			&summary.TotalPayments,
			&summary.ChargesCount,
			&summary.PaymentsCount,
		)

		if err != nil {
			return nil, err
		}

		summary.Balance = summary.TotalPayments - summary.TotalCharges
		summaries[summary.ApartmentID] = summary
	}

	return summaries, rows.Err()
}

// CalculateBalance розраховує поточний баланс квартири.
func (r *SQLitePaymentRepository) CalculateBalance(apartmentID int) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) as balance
		FROM payments
		WHERE apartment_id = ?
	`

	var balance float64
	err := r.db.QueryRow(query, apartmentID).Scan(&balance)
	return balance, err
}

// CalculateBalanceForPeriod розраховує баланс за період.
func (r *SQLitePaymentRepository) CalculateBalanceForPeriod(apartmentID int, period string) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) as balance
		FROM payments
		WHERE apartment_id = ? AND period = ?
	`

	var balance float64
	err := r.db.QueryRow(query, apartmentID, period).Scan(&balance)
	return balance, err
}

// scanPayments - допоміжна функція для сканування платежів.
func (r *SQLitePaymentRepository) scanPayments(rows *sql.Rows) ([]*domain.Payment, error) {
	var payments []*domain.Payment

	for rows.Next() {
		payment := &domain.Payment{}
		var notes sql.NullString

		err := rows.Scan(
			&payment.ID,
			&payment.ApartmentID,
			&payment.Type,
			&payment.Category,
			&payment.Amount,
			&payment.Description,
			&payment.PaymentDate,
			&payment.Period,
			&payment.CreatedBy,
			&notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if notes.Valid {
			payment.Notes = notes.String
		}

		payments = append(payments, payment)
	}

	return payments, rows.Err()
}

// GetMonthlyReport генерує місячний звіт.
func (r *SQLitePaymentRepository) GetMonthlyReport(apartmentID int, period string) (*domain.MonthlyReport, error) {
	// Отримуємо дані за поточний період
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) as total_charges,
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) as total_payments
		FROM payments
		WHERE apartment_id = ? AND period = ?
	`

	report := &domain.MonthlyReport{
		ApartmentID: apartmentID,
		Period:      period,
		ByCategory:  make(map[domain.PaymentCategory]float64),
	}

	err := r.db.QueryRow(query, apartmentID, period).Scan(
		&report.TotalCharges,
		&report.TotalPayments,
	)

	if err != nil {
		return nil, err
	}

	report.Balance = report.TotalPayments - report.TotalCharges

	// Отримуємо розбивку по категоріях
	categoryQuery := `
		SELECT category, SUM(amount) as total
		FROM payments
		WHERE apartment_id = ? AND period = ? AND type = 'charge'
		GROUP BY category
	`

	rows, err := r.db.Query(categoryQuery, apartmentID, period)
	if err != nil {
		return report, nil // Повертаємо без розбивки
	}
	defer rows.Close()

	for rows.Next() {
		var category domain.PaymentCategory
		var total float64

		if err := rows.Scan(&category, &total); err != nil {
			continue
		}

		report.ByCategory[category] = total
	}

	return report, nil
}

// GetDebtorApartments повертає ID квартир з боргом.
func (r *SQLitePaymentRepository) GetDebtorApartments() ([]int, error) {
	query := `
		SELECT apartment_id
		FROM payments
		GROUP BY apartment_id
		HAVING 
			COALESCE(SUM(CASE WHEN type = 'incoming' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'charge' THEN amount ELSE 0 END), 0) < 0
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debtors []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		debtors = append(debtors, id)
	}

	return debtors, rows.Err()
}

// GetTotalCharges повертає загальну суму нарахувань.
func (r *SQLitePaymentRepository) GetTotalCharges(apartmentID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE type = 'charge'
	`
	args := []interface{}{}

	if apartmentID > 0 {
		query += ` AND apartment_id = ?`
		args = append(args, apartmentID)
	}

	var total float64
	err := r.db.QueryRow(query, args...).Scan(&total)
	return total, err
}

// GetTotalPayments повертає загальну суму платежів.
func (r *SQLitePaymentRepository) GetTotalPayments(apartmentID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE type = 'incoming'
	`
	args := []interface{}{}

	if apartmentID > 0 {
		query += ` AND apartment_id = ?`
		args = append(args, apartmentID)
	}

	var total float64
	err := r.db.QueryRow(query, args...).Scan(&total)
	return total, err
}

// GetPaymentsByCategory повертає суми по категоріях.
func (r *SQLitePaymentRepository) GetPaymentsByCategory(apartmentID int, period string) (map[domain.PaymentCategory]float64, error) {
	query := `
		SELECT category, SUM(amount) as total
		FROM payments
		WHERE type = 'charge'
	`
	args := []interface{}{}

	if apartmentID > 0 {
		query += ` AND apartment_id = ?`
		args = append(args, apartmentID)
	}

	if period != "" {
		query += ` AND period = ?`
		args = append(args, period)
	}

	query += ` GROUP BY category`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[domain.PaymentCategory]float64)

	for rows.Next() {
		var category domain.PaymentCategory
		var total float64

		if err := rows.Scan(&category, &total); err != nil {
			return nil, err
		}

		result[category] = total
	}

	return result, rows.Err()
}

// GetRecentPayments повертає останні платежі.
func (r *SQLitePaymentRepository) GetRecentPayments(limit int) ([]*domain.Payment, error) {
	query := `
		SELECT id, apartment_id, type, category, amount, description,
		       payment_date, period, created_by, notes,
		       created_at, updated_at
		FROM payments
		ORDER BY payment_date DESC, id DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPayments(rows)
}

// Count підраховує кількість платежів.
func (r *SQLitePaymentRepository) Count(apartmentID int) (int, error) {
	query := `SELECT COUNT(*) FROM payments`
	args := []interface{}{}

	if apartmentID > 0 {
		query += ` WHERE apartment_id = ?`
		args = append(args, apartmentID)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}
