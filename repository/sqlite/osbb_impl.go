package sqlite

import (
	"database/sql"
	"fmt"
	"osbb-accounting/domain"
	"time"
)

// OSBBRepositoryImpl реалізує repository.OSBBRepository для SQLite.
type OSBBRepositoryImpl struct {
	db *sql.DB
}

// NewOSBBRepository створює новий екземпляр OSBBRepositoryImpl.
//
// Параметри:
//   - db: з'єднання з базою даних SQLite
//
// Повертає:
//   - *OSBBRepositoryImpl: новий репозиторій
func NewOSBBRepository(db *sql.DB) *OSBBRepositoryImpl {
	return &OSBBRepositoryImpl{db: db}
}

// Save зберігає дані ОСББ в БД.
// В системі може існувати тільки одна організація ОСББ (id=1).
func (r *OSBBRepositoryImpl) Save(osbb *domain.OSBB) error {
	// Перевіряємо валідність даних
	if err := osbb.Validate(); err != nil {
		return err
	}

	// Перевіряємо, чи вже існує ОСББ
	exists, err := r.Exists()
	if err != nil {
		return fmt.Errorf("помилка перевірки існування ОСББ: %w", err)
	}
	if exists {
		return domain.ErrOSBBAlreadyExists
	}

	// Встановлюємо ID=1 (єдина організація)
	osbb.ID = 1
	osbb.CreatedAt = time.Now()
	osbb.UpdatedAt = time.Now()

	query := `
		INSERT INTO osbb (
			id, name, short_name, address, edrpou, base_rate,
			chairman_name, chairman_phone, chairman_email,
			bank_name, bank_account, mfo, founded_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		osbb.ID,
		osbb.Name,
		osbb.ShortName,
		osbb.Address,
		osbb.EDRPOU,
		osbb.BaseRate,
		osbb.ChairmanName,
		osbb.ChairmanPhone,
		osbb.ChairmanEmail,
		osbb.BankName,
		osbb.BankAccount,
		osbb.MFO,
		osbb.FoundedAt,
		osbb.CreatedAt,
		osbb.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("помилка збереження ОСББ: %w", err)
	}

	return nil
}

// Get отримує дані ОСББ з БД.
func (r *OSBBRepositoryImpl) Get() (*domain.OSBB, error) {
	query := `
		SELECT 
			id, name, short_name, address, edrpou, base_rate,
			chairman_name, chairman_phone, chairman_email,
			bank_name, bank_account, mfo, founded_at,
			created_at, updated_at
		FROM osbb
		WHERE id = 1
	`

	osbb := &domain.OSBB{}
	var shortName, chairmanEmail sql.NullString

	err := r.db.QueryRow(query).Scan(
		&osbb.ID,
		&osbb.Name,
		&shortName,
		&osbb.Address,
		&osbb.EDRPOU,
		&osbb.BaseRate,
		&osbb.ChairmanName,
		&osbb.ChairmanPhone,
		&chairmanEmail,
		&osbb.BankName,
		&osbb.BankAccount,
		&osbb.MFO,
		&osbb.FoundedAt,
		&osbb.CreatedAt,
		&osbb.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrOSBBNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка отримання ОСББ: %w", err)
	}

	// Обробка nullable полів
	if shortName.Valid {
		osbb.ShortName = shortName.String
	}
	if chairmanEmail.Valid {
		osbb.ChairmanEmail = chairmanEmail.String
	}

	return osbb, nil
}

// Update оновлює дані ОСББ.
func (r *OSBBRepositoryImpl) Update(osbb *domain.OSBB) error {
	// Перевіряємо валідність даних
	if err := osbb.Validate(); err != nil {
		return err
	}

	// ID завжди має бути 1
	osbb.ID = 1
	osbb.UpdatedAt = time.Now()

	query := `
		UPDATE osbb SET
			name = ?,
			short_name = ?,
			address = ?,
			edrpou = ?,
			base_rate = ?,
			chairman_name = ?,
			chairman_phone = ?,
			chairman_email = ?,
			bank_name = ?,
			bank_account = ?,
			mfo = ?,
			founded_at = ?,
			updated_at = ?
		WHERE id = 1
	`

	result, err := r.db.Exec(
		query,
		osbb.Name,
		osbb.ShortName,
		osbb.Address,
		osbb.EDRPOU,
		osbb.BaseRate,
		osbb.ChairmanName,
		osbb.ChairmanPhone,
		osbb.ChairmanEmail,
		osbb.BankName,
		osbb.BankAccount,
		osbb.MFO,
		osbb.FoundedAt,
		osbb.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("помилка оновлення ОСББ: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату оновлення: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOSBBNotFound
	}

	return nil
}

// Exists перевіряє, чи існує ОСББ в системі.
func (r *OSBBRepositoryImpl) Exists() (bool, error) {
	query := `SELECT COUNT(*) FROM osbb WHERE id = 1`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("помилка перевірки існування ОСББ: %w", err)
	}

	return count > 0, nil
}

// GetSettings отримує налаштування та статистику ОСББ.
func (r *OSBBRepositoryImpl) GetSettings() (*domain.OSBBSettings, error) {
	// Отримуємо дані ОСББ
	osbb, err := r.Get()
	if err != nil {
		return nil, err
	}

	settings := &domain.OSBBSettings{
		OSBB: osbb,
	}

	// Отримуємо загальну кількість квартир
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM apartments WHERE is_active = 1
	`).Scan(&settings.TotalApartments)
	if err != nil {
		return nil, fmt.Errorf("помилка підрахунку квартир: %w", err)
	}

	// Отримуємо загальну площу
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(total_area), 0) FROM apartments WHERE is_active = 1
	`).Scan(&settings.TotalArea)
	if err != nil {
		return nil, fmt.Errorf("помилка підрахунку площі: %w", err)
	}

	// Отримуємо кількість власників
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM owners WHERE is_active = 1
	`).Scan(&settings.TotalOwners)
	if err != nil {
		return nil, fmt.Errorf("помилка підрахунку власників: %w", err)
	}

	// Отримуємо кількість активних особистих рахунків
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM personal_accounts WHERE is_active = 1
	`).Scan(&settings.ActiveAccounts)
	if err != nil {
		return nil, fmt.Errorf("помилка підрахунку рахунків: %w", err)
	}

	return settings, nil
}
