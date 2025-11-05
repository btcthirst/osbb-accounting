package sqlite

import (
	"database/sql"
	"fmt"
	"osbb-accounting/domain"
	"strings"
	"time"
)

// OwnerRepositoryImpl реалізує repository.OwnerRepository для SQLite.
type OwnerRepositoryImpl struct {
	db *sql.DB
}

// NewOwnerRepository створює новий екземпляр OwnerRepositoryImpl.
func NewOwnerRepository(db *sql.DB) *OwnerRepositoryImpl {
	return &OwnerRepositoryImpl{db: db}
}

// Save зберігає нового власника в БД.
func (r *OwnerRepositoryImpl) Save(owner *domain.Owner) error {
	// Перевіряємо валідність даних
	if err := owner.Validate(); err != nil {
		return err
	}

	// Перевіряємо унікальність ІПН
	existingOwner, err := r.FindByTaxID(owner.TaxID)
	if err == nil && existingOwner != nil {
		return domain.ErrOwnerAlreadyExists
	}

	owner.CreatedAt = time.Now()
	owner.UpdatedAt = time.Now()
	owner.IsActive = true

	query := `
		INSERT INTO owners (
			full_name, tax_id, phone, email, alternative_phone,
			passport_series, passport_number, notes, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		owner.FullName,
		owner.TaxID,
		owner.Phone,
		nullString(owner.Email),
		nullString(owner.AlternativePhone),
		nullString(owner.PassportSeries),
		nullString(owner.PassportNumber),
		nullString(owner.Notes),
		owner.IsActive,
		owner.CreatedAt,
		owner.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("помилка збереження власника: %w", err)
	}

	// Отримуємо ID створеного власника
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("помилка отримання ID власника: %w", err)
	}
	owner.ID = int(id)

	return nil
}

// FindByID знаходить власника за ID.
func (r *OwnerRepositoryImpl) FindByID(id int) (*domain.Owner, error) {
	query := `
		SELECT 
			id, full_name, tax_id, phone, email, alternative_phone,
			passport_series, passport_number, notes, is_active,
			created_at, updated_at
		FROM owners
		WHERE id = ?
	`

	owner := &domain.Owner{}
	var email, altPhone, passportSeries, passportNumber, notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&owner.ID,
		&owner.FullName,
		&owner.TaxID,
		&owner.Phone,
		&email,
		&altPhone,
		&passportSeries,
		&passportNumber,
		&notes,
		&owner.IsActive,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrOwnerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку власника: %w", err)
	}

	// Обробка nullable полів
	owner.Email = email.String
	owner.AlternativePhone = altPhone.String
	owner.PassportSeries = passportSeries.String
	owner.PassportNumber = passportNumber.String
	owner.Notes = notes.String

	return owner, nil
}

// FindByTaxID знаходить власника за ІПН.
func (r *OwnerRepositoryImpl) FindByTaxID(taxID string) (*domain.Owner, error) {
	query := `
		SELECT 
			id, full_name, tax_id, phone, email, alternative_phone,
			passport_series, passport_number, notes, is_active,
			created_at, updated_at
		FROM owners
		WHERE tax_id = ?
	`

	owner := &domain.Owner{}
	var email, altPhone, passportSeries, passportNumber, notes sql.NullString

	err := r.db.QueryRow(query, taxID).Scan(
		&owner.ID,
		&owner.FullName,
		&owner.TaxID,
		&owner.Phone,
		&email,
		&altPhone,
		&passportSeries,
		&passportNumber,
		&notes,
		&owner.IsActive,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrOwnerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку власника за ІПН: %w", err)
	}

	// Обробка nullable полів
	owner.Email = email.String
	owner.AlternativePhone = altPhone.String
	owner.PassportSeries = passportSeries.String
	owner.PassportNumber = passportNumber.String
	owner.Notes = notes.String

	return owner, nil
}

// FindAll повертає список всіх активних власників.
func (r *OwnerRepositoryImpl) FindAll() ([]*domain.Owner, error) {
	query := `
		SELECT 
			id, full_name, tax_id, phone, email, alternative_phone,
			passport_series, passport_number, notes, is_active,
			created_at, updated_at
		FROM owners
		WHERE is_active = 1
		ORDER BY full_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання списку власників: %w", err)
	}
	defer rows.Close()

	var owners []*domain.Owner
	for rows.Next() {
		owner := &domain.Owner{}
		var email, altPhone, passportSeries, passportNumber, notes sql.NullString

		err := rows.Scan(
			&owner.ID,
			&owner.FullName,
			&owner.TaxID,
			&owner.Phone,
			&email,
			&altPhone,
			&passportSeries,
			&passportNumber,
			&notes,
			&owner.IsActive,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("помилка сканування власника: %w", err)
		}

		// Обробка nullable полів
		owner.Email = email.String
		owner.AlternativePhone = altPhone.String
		owner.PassportSeries = passportSeries.String
		owner.PassportNumber = passportNumber.String
		owner.Notes = notes.String

		owners = append(owners, owner)
	}

	return owners, nil
}

// Search шукає власників за ПІБ або ІПН.
func (r *OwnerRepositoryImpl) Search(query string) ([]*domain.Owner, error) {
	searchQuery := `
		SELECT 
			id, full_name, tax_id, phone, email, alternative_phone,
			passport_series, passport_number, notes, is_active,
			created_at, updated_at
		FROM owners
		WHERE is_active = 1
		  AND (full_name LIKE ? OR tax_id LIKE ?)
		ORDER BY full_name
	`

	searchPattern := "%" + strings.ToLower(query) + "%"

	rows, err := r.db.Query(searchQuery, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку власників: %w", err)
	}
	defer rows.Close()

	var owners []*domain.Owner
	for rows.Next() {
		owner := &domain.Owner{}
		var email, altPhone, passportSeries, passportNumber, notes sql.NullString

		err := rows.Scan(
			&owner.ID,
			&owner.FullName,
			&owner.TaxID,
			&owner.Phone,
			&email,
			&altPhone,
			&passportSeries,
			&passportNumber,
			&notes,
			&owner.IsActive,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("помилка сканування власника: %w", err)
		}

		// Обробка nullable полів
		owner.Email = email.String
		owner.AlternativePhone = altPhone.String
		owner.PassportSeries = passportSeries.String
		owner.PassportNumber = passportNumber.String
		owner.Notes = notes.String

		owners = append(owners, owner)
	}

	return owners, nil
}

// Update оновлює дані власника.
func (r *OwnerRepositoryImpl) Update(owner *domain.Owner) error {
	// Перевіряємо валідність даних
	if err := owner.Validate(); err != nil {
		return err
	}

	owner.UpdatedAt = time.Now()

	query := `
		UPDATE owners SET
			full_name = ?,
			tax_id = ?,
			phone = ?,
			email = ?,
			alternative_phone = ?,
			passport_series = ?,
			passport_number = ?,
			notes = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		owner.FullName,
		owner.TaxID,
		owner.Phone,
		nullString(owner.Email),
		nullString(owner.AlternativePhone),
		nullString(owner.PassportSeries),
		nullString(owner.PassportNumber),
		nullString(owner.Notes),
		owner.UpdatedAt,
		owner.ID,
	)

	if err != nil {
		return fmt.Errorf("помилка оновлення власника: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату оновлення: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Deactivate деактивує власника (м'яке видалення).
func (r *OwnerRepositoryImpl) Deactivate(id int) error {
	query := `UPDATE owners SET is_active = 0, updated_at = ? WHERE id = ?`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("помилка деактивації власника: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату деактивації: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// Activate активує власника.
func (r *OwnerRepositoryImpl) Activate(id int) error {
	query := `UPDATE owners SET is_active = 1, updated_at = ? WHERE id = ?`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("помилка активації власника: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату активації: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOwnerNotFound
	}

	return nil
}

// GetWithAccounts отримує власника з його особистими рахунками.
func (r *OwnerRepositoryImpl) GetWithAccounts(ownerID int) (*domain.OwnerWithAccounts, error) {
	// Отримуємо власника
	owner, err := r.FindByID(ownerID)
	if err != nil {
		return nil, err
	}

	// Отримуємо рахунки власника
	accountQuery := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE owner_id = ?
		ORDER BY account_number
	`

	rows, err := r.db.Query(accountQuery, ownerID)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання рахунків власника: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.PersonalAccount
	for rows.Next() {
		account := &domain.PersonalAccount{}
		var closedAt sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&account.ID,
			&account.AccountNumber,
			&account.ApartmentID,
			&account.OwnerID,
			&account.OpenedAt,
			&closedAt,
			&account.CurrentBalance,
			&account.IsActive,
			&notes,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("помилка сканування рахунку: %w", err)
		}

		if closedAt.Valid {
			account.ClosedAt = &closedAt.Time
		}
		account.Notes = notes.String

		accounts = append(accounts, account)
	}

	return &domain.OwnerWithAccounts{
		Owner:    owner,
		Accounts: accounts,
	}, nil
}

// Count підраховує кількість активних власників.
func (r *OwnerRepositoryImpl) Count() (int, error) {
	query := `SELECT COUNT(*) FROM owners WHERE is_active = 1`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("помилка підрахунку власників: %w", err)
	}

	return count, nil
}

// nullString повертає sql.NullString для обробки NULL значень.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
