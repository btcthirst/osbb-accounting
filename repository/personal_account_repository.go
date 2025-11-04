package repository

import "osbb-accounting/domain"

// =============================================================================
// PersonalAccountRepository - управління особистими рахунками
// =============================================================================

// PersonalAccountRepository визначає контракт для операцій з особистими рахунками.
type PersonalAccountRepository interface {
	// Save створює новий особистий рахунок.
	//
	// Параметри:
	//   - account: дані рахунку
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrAccountAlreadyExists якщо номер зайнятий
	Save(account *domain.PersonalAccount) error

	// FindByID знаходить рахунок за ID.
	//
	// Параметри:
	//   - id: ID рахунку
	//
	// Повертає:
	//   - *domain.PersonalAccount: знайдений рахунок
	//   - error: domain.ErrAccountNotFound якщо не знайдено
	FindByID(id int) (*domain.PersonalAccount, error)

	// FindByAccountNumber знаходить рахунок за номером.
	//
	// Параметри:
	//   - accountNumber: номер особистого рахунку
	//
	// Повертає:
	//   - *domain.PersonalAccount: знайдений рахунок
	//   - error: domain.ErrAccountNotFound якщо не знайдено
	FindByAccountNumber(accountNumber string) (*domain.PersonalAccount, error)

	// FindByApartmentID знаходить рахунок квартири.
	//
	// Параметри:
	//   - apartmentID: ID квартири
	//
	// Повертає:
	//   - *domain.PersonalAccount: рахунок квартири
	//   - error: domain.ErrAccountNotFound якщо не знайдено
	FindByApartmentID(apartmentID int) (*domain.PersonalAccount, error)

	// FindByOwnerID знаходить всі рахунки власника.
	//
	// Параметри:
	//   - ownerID: ID власника
	//
	// Повертає:
	//   - []*domain.PersonalAccount: список рахунків
	//   - error: nil при успіху
	FindByOwnerID(ownerID int) ([]*domain.PersonalAccount, error)

	// FindAll повертає список всіх активних рахунків.
	//
	// Повертає:
	//   - []*domain.PersonalAccount: список рахунків
	//   - error: nil при успіху
	FindAll() ([]*domain.PersonalAccount, error)

	// FindWithDebt повертає рахунки з заборгованістю.
	//
	// Повертає:
	//   - []*domain.PersonalAccount: рахунки з боргом
	//   - error: nil при успіху
	FindWithDebt() ([]*domain.PersonalAccount, error)

	// Update оновлює дані рахунку.
	//
	// Параметри:
	//   - account: оновлені дані
	//
	// Повертає:
	//   - error: nil при успіху
	Update(account *domain.PersonalAccount) error

	// UpdateBalance оновлює баланс рахунку.
	//
	// Параметри:
	//   - accountID: ID рахунку
	//   - newBalance: новий баланс
	//
	// Повертає:
	//   - error: nil при успіху
	UpdateBalance(accountID int, newBalance float64) error

	// Close закриває особистий рахунок.
	// Рахунок не можна закрити, якщо є непогашена заборгованість.
	//
	// Параметри:
	//   - accountID: ID рахунку
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrCannotCloseAccountWithDebt якщо є борг
	Close(accountID int) error

	// Reopen знову відкриває закритий рахунок.
	//
	// Параметри:
	//   - accountID: ID рахунку
	//
	// Повертає:
	//   - error: nil при успіху
	Reopen(accountID int) error

	// GetTotalDebt повертає загальну суму боргу по всіх рахунках.
	//
	// Повертає:
	//   - float64: загальний борг
	//   - error: nil при успіху
	GetTotalDebt() (float64, error)

	// GetTotalOverpayment повертає загальну суму переплат.
	//
	// Повертає:
	//   - float64: загальна переплата
	//   - error: nil при успіху
	GetTotalOverpayment() (float64, error)

	// Count підраховує кількість активних рахунків.
	//
	// Повертає:
	//   - int: кількість рахунків
	//   - error: nil при успіху
	Count() (int, error)

	// GenerateAccountNumber генерує унікальний номер особистого рахунку.
	//
	// Повертає:
	//   - string: згенерований номер рахунку
	//   - error: nil при успіху
	GenerateAccountNumber() (string, error)
}
