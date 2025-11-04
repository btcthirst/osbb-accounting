package sqlite

import (
	"database/sql"
	"errors"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"
)

// SQLiteOwnerRepository - SQLite реалізація OwnerRepository.
type SQLiteOwnerRepository struct {
	db *sql.DB
}

// NewSQLiteOwnerRepository створює новий екземпляр репозиторію власників.
func NewSQLiteOwnerRepository(db *sql.DB) repository.OwnerRepository {
	return &SQLiteOwnerRepository{db: db}
}

// Save зберігає нового власника.
func (r *SQLiteOwnerRepository) Save(owner *domain.Owner) error {
	if err := owner.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO owners (
			full_name, tax_id, phone, phone_additional, email,
			passport_series, passport_number, passport_issued_by, passport_issued_date,
			registration_address, notes, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	owner.CreatedAt = now
	owner.UpdatedAt = now
	owner.IsActive = true

	result, err := r.db.Exec(
		query,
		owner.FullName,
		owner.TaxID,
		owner.Phone,
		owner.PhoneAdditional,
		owner.Email,
		owner.PassportSeries,
		owner.PassportNumber,
		owner.PassportIssuedBy,
		owner.PassportIssuedDate,
		owner.RegistrationAddress,
		owner.Notes,
		owner.IsActive,
		now,
		now,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrOwnerAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	owner.ID = int(id)
	return nil
}

// FindByID знаходить власника за ID.
func (r *SQLiteOwnerRepository) FindByID(id int) (*domain.Owner, error) {
	query := `
		SELECT id, full_name, tax_id, phone, phone_additional, email,
		       passport_series, passport_number, passport_issued_by, passport_issued_date,
		       registration_address, notes, is_active, created_at, updated_at
		FROM owners
		WHERE id = ? AND is_active = 1
	`

	owner := &domain.Owner{}
	var phoneAdditional, email, passportSeries, passportNumber sql.NullString
	var passportIssuedBy, registrationAddress, notes sql.NullString
	var passportIssuedDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&owner.ID,
		&owner.FullName,
		&owner.TaxID,
		&owner.Phone,
		&phoneAdditional,
		&email,
		&passportSeries,
		&passportNumber,
		&passportIssuedBy,
		&passportIssuedDate,
		&registrationAddress,
		&notes,
		&owner.IsActive,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOwnerNotFound
		}
		return nil, err
	}

	// Конвертуємо nullable поля
	if phoneAdditional.Valid {
		owner.PhoneAdditional = phoneAdditional.String
	}
	if email.Valid {
		owner.Email = email.String
	}
	if passportSeries.Valid {
		owner.PassportSeries = passportSeries.String
	}
	if passportNumber.Valid {
		owner.PassportNumber = passportNumber.String
	}
	if passportIssuedBy.Valid {
		owner.PassportIssuedBy = passportIssuedBy.String
	}
	if passportIssuedDate.Valid {
		owner.PassportIssuedDate = &passportIssuedDate.Time
	}
	if registrationAddress.Valid {
		owner.RegistrationAddress = registrationAddress.String
	}
	if notes.Valid {
		owner.Notes = notes.String
	}

	return owner, nil
}

// FindByTaxID знаходить власника за ІПН.
func (r *SQLiteOwnerRepository) FindByTaxID(taxID string) (*domain.Owner, error) {
	query := `
		SELECT id, full_name, tax_id, phone, phone_additional, email,
		       passport_series, passport_number, passport_issued_by, passport_issued_date,
		       registration_address, notes, is_active, created_at, updated_at
		FROM owners
		WHERE tax_id = ? AND is_active = 1
	`

	owner := &domain.Owner{}
	var phoneAdditional, email, passportSeries, passportNumber sql.NullString
	var passportIssuedBy, registrationAddress, notes sql.NullString
	var passportIssuedDate sql.NullTime

	err := r.db.QueryRow(query, taxID).Scan(
		&owner.ID,
		&owner.FullName,
		&owner.TaxID,
		&owner.Phone,
		&phoneAdditional,
		&email,
		&passportSeries,
		&passportNumber,
		&passportIssuedBy,
		&passportIssuedDate,
		&registrationAddress,
		&notes,
		&owner.IsActive,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOwnerNotFound
		}
		return nil, err
	}

	// Конвертуємо nullable поля (аналогічно FindByID)
	if phoneAdditional.Valid {
		owner.PhoneAdditional = phoneAdditional.String
	}
	if email.Valid {
		owner.Email = email.String
	}
	if notes.Valid {
		owner.Notes = notes.String
	}

	return owner, nil
}

// FindAll повертає всіх активних власників.
func (r *SQLiteOwnerRepository) FindAll() ([]*domain.Owner, error) {
	query := `
		SELECT id, full_name, tax_id, phone, phone_additional, email,
		       passport_series, passport_number, passport_issued_by, passport_issued_date,
		       registration_address, notes, is_active, created_at, updated_at
		FROM owners
		WHERE is_active = 1
		ORDER BY full_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOwners(rows)
}

// Search шукає власників за ПІБ або телефоном.
func (r *SQLiteOwnerRepository) Search(query string) ([]*domain.Owner, error) {
	sqlQuery := `
		SELECT id, full_name, tax_id, phone, phone_additional, email,
		       passport_series, passport_number, passport_issued_by, passport_issued_date,
		       registration_address, notes, is_active, created_at, updated_at
		FROM owners
		WHERE is_active = 1 AND (full_name LIKE ? OR phone LIKE ? OR tax_id LIKE ?)
		ORDER BY full_name
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(sqlQuery, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOwners(rows)
}

// Update оновлює дані власника.
func (r *SQLiteOwnerRepository) Update(owner *domain.Owner) error {
	if err := owner.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE owners
		SET full_name = ?, tax_id = ?, phone = ?, phone_additional = ?, email = ?,
		    passport_series = ?, passport_number = ?, passport_issued_by = ?,
		    passport_issued_date = ?, registration_address = ?, notes = ?,
		    updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	owner.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		owner.FullName,
		owner.TaxID,
		owner.Phone,
		owner.AlternativePhone,
		owner.Email,
		owner.PassportSeries,
		owner.PassportNumber,
		owner.PassportIssuedBy,
		owner.PassportIssuedDate,
		owner.RegistrationAddress,
		owner.Notes,
		owner.UpdatedAt,
		owner.ID,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrOwnerAlreadyExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Deactivate деактивує власника.
func (r *SQLiteOwnerRepository) Deactivate(id int) error {
	query := `UPDATE owners SET is_active = 0, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Activate активує власника.
func (r *SQLiteOwnerRepository) Activate(id int) error {
	query := `UPDATE owners SET is_active = 1, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Delete видаляє власника.
func (r *SQLiteOwnerRepository) Delete(id int) error {
	query := `DELETE FROM owners WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Count підраховує кількість власників.
func (r *SQLiteOwnerRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM owners WHERE is_active = 1").Scan(&count)
	return count, err
}

// GetOwnersWithApartments повертає мапу власник->кількість квартир.
func (r *SQLiteOwnerRepository) GetOwnersWithApartments() (map[int]int, error) {
	query := `
		SELECT owner_id, COUNT(*) as apartment_count
		FROM personal_accounts
		WHERE is_active = 1
		GROUP BY owner_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var ownerID, count int
		if err := rows.Scan(&ownerID, &count); err != nil {
			return nil, err
		}
		result[ownerID] = count
	}

	return result, rows.Err()
}

// scanOwners - допоміжна функція для сканування власників.
func (r *SQLiteOwnerRepository) scanOwners(rows *sql.Rows) ([]*domain.Owner, error) {
	var owners []*domain.Owner

	for rows.Next() {
		owner := &domain.Owner{}
		var phoneAdditional, email, passportSeries, passportNumber sql.NullString
		var passportIssuedBy, registrationAddress, notes sql.NullString
		var passportIssuedDate sql.NullTime

		err := rows.Scan(
			&owner.ID,
			&owner.FullName,
			&owner.TaxID,
			&owner.Phone,
			&phoneAdditional,
			&email,
			&passportSeries,
			&passportNumber,
			&passportIssuedBy,
			&passportIssuedDate,
			&registrationAddress,
			&notes,
			&owner.IsActive,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Конвертуємо nullable поля
		if phoneAdditional.Valid {
			owner.AlternativePhone = phoneAdditional.String
		}
		if email.Valid {
			owner.Email = email.String
		}
		if passportSeries.Valid {
			owner.PassportSeries = passportSeries.String
		}
		if passportNumber.Valid {
			owner.PassportNumber = passportNumber.String
		}
		if passportIssuedBy.Valid {
			owner.PassportIssuedBy = passportIssuedBy.String
		}
		if passportIssuedDate.Valid {
			owner.PassportIssuedDate = &passportIssuedDate.Time
		}
		if registrationAddress.Valid {
			owner.RegistrationAddress = registrationAddress.String
		}
		if notes.Valid {
			owner.Notes = notes.String
		}

		owners = append(owners, owner)
	}

	return owners, rows.Err()
}
