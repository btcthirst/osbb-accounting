package sqlite

import (
	"database/sql"
	"fmt"
	"osbb-accounting/domain"
	"time"
)

// PersonalAccountRepositoryImpl реалізує repository.PersonalAccountRepository для SQLite.
type PersonalAccountRepositoryImpl struct {
	db *sql.DB
}

// NewPersonalAccountRepository створює новий екземпляр PersonalAccountRepositoryImpl.
func NewPersonalAccountRepository(db *sql.DB) *PersonalAccountRepositoryImpl {
	return &PersonalAccountRepositoryImpl{db: db}
}

// Save створює новий особистий рахунок.
func (r *PersonalAccountRepositoryImpl) Save(account *domain.PersonalAccount) error {
	// Перевіряємо валідність даних
	if err := account.Validate(); err != nil {
		return err
	}

	// Перевіряємо унікальність номера рахунку
	existing, err := r.FindByAccountNumber(account.AccountNumber)
	if err == nil && existing != nil {
		return domain.ErrAccountAlreadyExists
	}

	// Перевіряємо, чи квартира вже має рахунок
	existingByApt, err := r.FindByApartmentID(account.ApartmentID)
	if err == nil && existingByApt != nil {
		return domain.ErrAccountAlreadyAssigned
	}

	account.OpenedAt = time.Now()
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	account.IsActive = true
	account.CurrentBalance = 0

	query := `
		INSERT INTO personal_accounts (
			account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		account.AccountNumber,
		account.ApartmentID,
		account.OwnerID,
		account.OpenedAt,
		nil, // closed_at
		account.CurrentBalance,
		account.IsActive,
		nullString(account.Notes),
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("помилка збереження особистого рахунку: %w", err)
	}

	// Отримуємо ID створеного рахунку
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("помилка отримання ID рахунку: %w", err)
	}
	account.ID = int(id)

	return nil
}

// FindByID знаходить рахунок за ID.
func (r *PersonalAccountRepositoryImpl) FindByID(id int) (*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE id = ?
	`

	account := &domain.PersonalAccount{}
	var closedAt sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
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

	if err == sql.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку рахунку: %w", err)
	}

	// Обробка nullable полів
	if closedAt.Valid {
		account.ClosedAt = &closedAt.Time
	}
	account.Notes = notes.String

	return account, nil
}

// FindByAccountNumber знаходить рахунок за номером.
func (r *PersonalAccountRepositoryImpl) FindByAccountNumber(accountNumber string) (*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE account_number = ?
	`

	account := &domain.PersonalAccount{}
	var closedAt sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRow(query, accountNumber).Scan(
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

	if err == sql.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку рахунку за номером: %w", err)
	}

	// Обробка nullable полів
	if closedAt.Valid {
		account.ClosedAt = &closedAt.Time
	}
	account.Notes = notes.String

	return account, nil
}

// FindByApartmentID знаходить рахунок квартири.
func (r *PersonalAccountRepositoryImpl) FindByApartmentID(apartmentID int) (*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE apartment_id = ?
	`

	account := &domain.PersonalAccount{}
	var closedAt sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRow(query, apartmentID).Scan(
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

	if err == sql.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("помилка пошуку рахунку квартири: %w", err)
	}

	// Обробка nullable полів
	if closedAt.Valid {
		account.ClosedAt = &closedAt.Time
	}
	account.Notes = notes.String

	return account, nil
}

// FindByOwnerID знаходить всі рахунки власника.
func (r *PersonalAccountRepositoryImpl) FindByOwnerID(ownerID int) ([]*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE owner_id = ?
		ORDER BY account_number
	`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання рахунків власника: %w", err)
	}
	defer rows.Close()

	return r.scanAccounts(rows)
}

// FindAll повертає список всіх активних рахунків.
func (r *PersonalAccountRepositoryImpl) FindAll() ([]*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE is_active = 1
		ORDER BY account_number
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання списку рахунків: %w", err)
	}
	defer rows.Close()

	return r.scanAccounts(rows)
}

// FindWithDebt повертає рахунки з заборгованістю.
func (r *PersonalAccountRepositoryImpl) FindWithDebt() ([]*domain.PersonalAccount, error) {
	query := `
		SELECT 
			id, account_number, apartment_id, owner_id, opened_at, closed_at,
			current_balance, is_active, notes, created_at, updated_at
		FROM personal_accounts
		WHERE is_active = 1 AND current_balance < 0
		ORDER BY current_balance ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання рахунків з боргом: %w", err)
	}
	defer rows.Close()

	return r.scanAccounts(rows)
}

// Update оновлює дані рахунку.
func (r *PersonalAccountRepositoryImpl) Update(account *domain.PersonalAccount) error {
	// Перевіряємо валідність даних
	if err := account.Validate(); err != nil {
		return err
	}

	account.UpdatedAt = time.Now()

	query := `
		UPDATE personal_accounts SET
			account_number = ?,
			apartment_id = ?,
			owner_id = ?,
			current_balance = ?,
			notes = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		account.AccountNumber,
		account.ApartmentID,
		account.OwnerID,
		account.CurrentBalance,
		nullString(account.Notes),
		account.UpdatedAt,
		account.ID,
	)

	if err != nil {
		return fmt.Errorf("помилка оновлення рахунку: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату оновлення: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

// UpdateBalance оновлює баланс рахунку.
func (r *PersonalAccountRepositoryImpl) UpdateBalance(accountID int, newBalance float64) error {
	query := `
		UPDATE personal_accounts 
		SET current_balance = ?, updated_at = ? 
		WHERE id = ?
	`

	result, err := r.db.Exec(query, newBalance, time.Now(), accountID)
	if err != nil {
		return fmt.Errorf("помилка оновлення балансу: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату оновлення: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

// Close закриває особистий рахунок.
func (r *PersonalAccountRepositoryImpl) Close(accountID int) error {
	// Перевіряємо, чи рахунок існує та отримуємо його дані
	account, err := r.FindByID(accountID)
	if err != nil {
		return err
	}

	// Перевіряємо, чи немає боргу
	if account.HasDebt() {
		return domain.ErrCannotCloseAccountWithDebt
	}

	// Закриваємо рахунок
	now := time.Now()
	query := `
		UPDATE personal_accounts 
		SET closed_at = ?, is_active = 0, updated_at = ? 
		WHERE id = ?
	`

	_, err = r.db.Exec(query, now, now, accountID)
	if err != nil {
		return fmt.Errorf("помилка закриття рахунку: %w", err)
	}

	return nil
}

// Reopen знову відкриває закритий рахунок.
func (r *PersonalAccountRepositoryImpl) Reopen(accountID int) error {
	query := `
		UPDATE personal_accounts 
		SET closed_at = NULL, is_active = 1, updated_at = ? 
		WHERE id = ?
	`

	result, err := r.db.Exec(query, time.Now(), accountID)
	if err != nil {
		return fmt.Errorf("помилка відкриття рахунку: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("помилка перевірки результату відкриття: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

// GetTotalDebt повертає загальну суму боргу по всіх рахунках.
func (r *PersonalAccountRepositoryImpl) GetTotalDebt() (float64, error) {
	query := `
		SELECT COALESCE(SUM(ABS(current_balance)), 0) 
		FROM personal_accounts 
		WHERE is_active = 1 AND current_balance < 0
	`

	var totalDebt float64
	err := r.db.QueryRow(query).Scan(&totalDebt)
	if err != nil {
		return 0, fmt.Errorf("помилка підрахунку загального боргу: %w", err)
	}

	return totalDebt, nil
}

// GetTotalOverpayment повертає загальну суму переплат.
func (r *PersonalAccountRepositoryImpl) GetTotalOverpayment() (float64, error) {
	query := `
		SELECT COALESCE(SUM(current_balance), 0) 
		FROM personal_accounts 
		WHERE is_active = 1 AND current_balance > 0
	`

	var totalOverpayment float64
	err := r.db.QueryRow(query).Scan(&totalOverpayment)
	if err != nil {
		return 0, fmt.Errorf("помилка підрахунку загальної переплати: %w", err)
	}

	return totalOverpayment, nil
}

// Count підраховує кількість активних рахунків.
func (r *PersonalAccountRepositoryImpl) Count() (int, error) {
	query := `SELECT COUNT(*) FROM personal_accounts WHERE is_active = 1`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("помилка підрахунку рахунків: %w", err)
	}

	return count, nil
}

// GenerateAccountNumber генерує унікальний номер особистого рахунку.
func (r *PersonalAccountRepositoryImpl) GenerateAccountNumber() (string, error) {
	// Отримуємо максимальний номер рахунку
	query := `
		SELECT account_number 
		FROM personal_accounts 
		ORDER BY id DESC 
		LIMIT 1
	`

	var lastNumber string
	err := r.db.QueryRow(query).Scan(&lastNumber)
	if err == sql.ErrNoRows {
		// Перший рахунок
		return "0001-0001", nil
	}
	if err != nil {
		return "", fmt.Errorf("помилка генерації номера рахунку: %w", err)
	}

	// Парсимо останній номер та генеруємо наступний
	// Формат: XXXX-YYYY
	// Для простоти генеруємо послідовні номери
	var section1, section2 int
	_, err = fmt.Sscanf(lastNumber, "%04d-%04d", &section1, &section2)
	if err != nil {
		return "", fmt.Errorf("помилка парсингу номера рахунку: %w", err)
	}

	// Інкрементуємо
	section2++
	if section2 > 9999 {
		section2 = 1
		section1++
	}

	newNumber := fmt.Sprintf("%04d-%04d", section1, section2)
	return newNumber, nil
}

// scanAccounts - допоміжна функція для сканування рядків рахунків.
func (r *PersonalAccountRepositoryImpl) scanAccounts(rows *sql.Rows) ([]*domain.PersonalAccount, error) {
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

		// Обробка nullable полів
		if closedAt.Valid {
			account.ClosedAt = &closedAt.Time
		}
		account.Notes = notes.String

		accounts = append(accounts, account)
	}

	return accounts, nil
}
