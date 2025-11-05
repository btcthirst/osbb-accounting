package sqlite

import (
	"database/sql"
	"osbb-accounting/repository"
)

// RepositoryFactory створює всі репозиторії для SQLite.
type RepositoryFactory struct {
	db *sql.DB
}

// NewRepositoryFactory створює новий екземпляр фабрики репозиторіїв.
//
// Параметри:
//   - db: з'єднання з базою даних SQLite
//
// Повертає:
//   - *RepositoryFactory: фабрика репозиторіїв
func NewRepositoryFactory(db *sql.DB) *RepositoryFactory {
	return &RepositoryFactory{db: db}
}

// CreateUserRepository створює репозиторій користувачів.
func (f *RepositoryFactory) CreateUserRepository() repository.UserRepository {
	return NewSQLiteUserRepository(f.db)
}

// CreateApartmentRepository створює репозиторій квартир.
func (f *RepositoryFactory) CreateApartmentRepository() repository.ApartmentRepository {
	return NewSQLiteApartmentRepository(f.db)
}

// CreatePaymentRepository створює репозиторій платежів.
func (f *RepositoryFactory) CreatePaymentRepository() repository.PaymentRepository {
	return NewSQLitePaymentRepository(f.db)
}

// CreateOSBBRepository створює репозиторій ОСББ.
func (f *RepositoryFactory) CreateOSBBRepository() repository.OSBBRepository {
	return NewOSBBRepository(f.db)
}

// CreateOwnerRepository створює репозиторій власників.
func (f *RepositoryFactory) CreateOwnerRepository() repository.OwnerRepository {
	return NewOwnerRepository(f.db)
}

// CreatePersonalAccountRepository створює репозиторій особистих рахунків.
func (f *RepositoryFactory) CreatePersonalAccountRepository() repository.PersonalAccountRepository {
	return NewPersonalAccountRepository(f.db)
}

// CreateAllRepositories створює всі репозиторії одночасно.
// Зручно для ініціалізації всієї системи.
//
// Повертає структуру з усіма репозиторіями.
func (f *RepositoryFactory) CreateAllRepositories() *AllRepositories {
	return &AllRepositories{
		User:            f.CreateUserRepository(),
		Apartment:       f.CreateApartmentRepository(),
		Payment:         f.CreatePaymentRepository(),
		OSBB:            f.CreateOSBBRepository(),
		Owner:           f.CreateOwnerRepository(),
		PersonalAccount: f.CreatePersonalAccountRepository(),
	}
}

// AllRepositories містить всі репозиторії системи.
type AllRepositories struct {
	User            repository.UserRepository
	Apartment       repository.ApartmentRepository
	Payment         repository.PaymentRepository
	OSBB            repository.OSBBRepository
	Owner           repository.OwnerRepository
	PersonalAccount repository.PersonalAccountRepository
}
