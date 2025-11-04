package repository

import "osbb-accounting/domain"

// PersonalAccountRepository визначає контракт для роботи з особистими рахунками.
type PersonalAccountRepository interface {
	// Save зберігає новий особистий рахунок.
	Save(account *domain.PersonalAccount) error

	// FindByID знаходить рахунок за ID.
	FindByID(id int) (*domain.PersonalAccount, error)

	// FindByAccountNumber знаходить рахунок за номером.
	FindByAccountNumber(accountNumber string) (*domain.PersonalAccount, error)

	// FindByApartmentID знаходить активний рахунок квартири.
	FindByApartmentID(apartmentID int) (*domain.PersonalAccount, error)

	// FindByOwnerID знаходить всі рахунки власника.
	FindByOwnerID(ownerID int) ([]*domain.PersonalAccount, error)

	// FindAll повертає всі активні рахунки.
	FindAll() ([]*domain.PersonalAccount, error)

	// FindWithDetails повертає рахунок з повною інформацією.
	FindWithDetails(id int) (*domain.PersonalAccountWithDetails, error)

	// FindAllWithDetails повертає всі рахунки з повною інформацією.
	FindAllWithDetails() ([]*domain.PersonalAccountWithDetails, error)

	// Update оновлює дані рахунку.
	Update(account *domain.PersonalAccount) error

	// Close закриває рахунок (встановлює close_date).
	Close(id int) error

	// Reopen відкриває закритий рахунок.
	Reopen(id int) error

	// Delete видаляє рахунок.
	Delete(id int) error

	// Count підраховує кількість рахунків.
	Count() (int, error)

	// GenerateAccountNumber генерує наступний вільний номер ОР.
	GenerateAccountNumber() (string, error)
}
