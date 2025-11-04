package repository

import "osbb-accounting/domain"

// =============================================================================
// OwnerRepository - управління власниками квартир
// =============================================================================

// OwnerRepository визначає контракт для операцій з власниками.
type OwnerRepository interface {
	// Save зберігає нового власника в БД.
	//
	// Параметри:
	//   - owner: дані власника
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrOwnerAlreadyExists якщо ІПН зайнятий
	Save(owner *domain.Owner) error

	// FindByID знаходить власника за ID.
	//
	// Параметри:
	//   - id: ID власника
	//
	// Повертає:
	//   - *domain.Owner: знайдений власник
	//   - error: domain.ErrOwnerNotFound якщо не знайдено
	FindByID(id int) (*domain.Owner, error)

	// FindByTaxID знаходить власника за ІПН.
	//
	// Параметри:
	//   - taxID: індивідуальний податковий номер
	//
	// Повертає:
	//   - *domain.Owner: знайдений власник
	//   - error: domain.ErrOwnerNotFound якщо не знайдено
	FindByTaxID(taxID string) (*domain.Owner, error)

	// FindAll повертає список всіх активних власників.
	//
	// Повертає:
	//   - []*domain.Owner: список власників
	//   - error: nil при успіху
	FindAll() ([]*domain.Owner, error)

	// Search шукає власників за ПІБ або ІПН.
	//
	// Параметри:
	//   - query: пошуковий запит
	//
	// Повертає:
	//   - []*domain.Owner: знайдені власники
	//   - error: nil при успіху
	Search(query string) ([]*domain.Owner, error)

	// Update оновлює дані власника.
	//
	// Параметри:
	//   - owner: оновлені дані
	//
	// Повертає:
	//   - error: nil при успіху
	Update(owner *domain.Owner) error

	// Deactivate деактивує власника (м'яке видалення).
	//
	// Параметри:
	//   - id: ID власника
	//
	// Повертає:
	//   - error: nil при успіху
	Deactivate(id int) error

	// Activate активує власника.
	//
	// Параметри:
	//   - id: ID власника
	//
	// Повертає:
	//   - error: nil при успіху
	Activate(id int) error

	// GetWithAccounts отримує власника з його особистими рахунками.
	//
	// Параметри:
	//   - ownerID: ID власника
	//
	// Повертає:
	//   - *domain.OwnerWithAccounts: власник з рахунками
	//   - error: nil при успіху
	GetWithAccounts(ownerID int) (*domain.OwnerWithAccounts, error)

	// Count підраховує кількість активних власників.
	//
	// Повертає:
	//   - int: кількість власників
	//   - error: nil при успіху
	Count() (int, error)
}
