package repository

import "osbb-accounting/domain"

// OwnerRepository визначає контракт для роботи з власниками.
type OwnerRepository interface {
	// Save зберігає нового власника.
	Save(owner *domain.Owner) error

	// FindByID знаходить власника за ID.
	FindByID(id int) (*domain.Owner, error)

	// FindByTaxID знаходить власника за ІПН.
	FindByTaxID(taxID string) (*domain.Owner, error)

	// FindAll повертає всіх активних власників.
	FindAll() ([]*domain.Owner, error)

	// Search шукає власників за ПІБ або телефоном.
	Search(query string) ([]*domain.Owner, error)

	// Update оновлює дані власника.
	Update(owner *domain.Owner) error

	// Deactivate деактивує власника.
	Deactivate(id int) error

	// Activate активує власника.
	Activate(id int) error

	// Delete видаляє власника (якщо немає прив'язаних рахунків).
	Delete(id int) error

	// Count підраховує кількість власників.
	Count() (int, error)

	// GetOwnersWithApartments повертає власників з кількістю квартир.
	GetOwnersWithApartments() (map[int]int, error)
}
