package repository

import "osbb-accounting/domain"

// ApartmentRepository визначає контракт для всіх операцій з квартирами.
// Дотримуємось тих самих принципів Repository Pattern, що і для User.
type ApartmentRepository interface {
	// Save зберігає нову квартиру в БД.
	// Автоматично генерує ID та встановлює CreatedAt, UpdatedAt.
	//
	// Параметри:
	//   - apartment: вказівник на domain.Apartment для збереження
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrApartmentAlreadyExists якщо номер зайнятий
	Save(apartment *domain.Apartment) error

	// FindByID знаходить квартиру за унікальним ідентифікатором.
	//
	// Параметри:
	//   - id: унікальний ідентифікатор квартири
	//
	// Повертає:
	//   - *domain.Apartment: знайдену квартиру
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не знайдено
	FindByID(id int) (*domain.Apartment, error)

	// FindByNumber знаходить квартиру за номером.
	//
	// Параметри:
	//   - apartmentNumber: номер квартири (унікальний)
	//
	// Повертає:
	//   - *domain.Apartment: знайдену квартиру
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не знайдено
	FindByNumber(apartmentNumber string) (*domain.Apartment, error)

	// FindAll повертає список всіх активних квартир.
	//
	// Повертає:
	//   - []*domain.Apartment: слайс квартир
	//   - error: nil при успіху, помилки при проблемах з БД
	FindAll() ([]*domain.Apartment, error)

	// FindByFilter знаходить квартири за вказаними критеріями.
	// Підтримує складні запити з множинними умовами та сортуванням.
	//
	// Параметри:
	//   - filter: параметри фільтрації та сортування
	//
	// Повертає:
	//   - []*domain.Apartment: слайс квартир, що відповідають критеріям
	//   - error: nil при успіху, помилки при проблемах з БД
	FindByFilter(filter *domain.ApartmentFilter) ([]*domain.Apartment, error)

	// FindByFloor знаходить всі квартири на вказаному поверсі.
	//
	// Параметри:
	//   - floor: номер поверху
	//
	// Повертає:
	//   - []*domain.Apartment: слайс квартир на поверсі
	//   - error: nil при успіху, помилки при проблемах з БД
	FindByFloor(floor int) ([]*domain.Apartment, error)

	// FindByEntrance знаходить всі квартири у вказаному під'їзді.
	//
	// Параметри:
	//   - entrance: номер під'їзду
	//
	// Повертає:
	//   - []*domain.Apartment: слайс квартир у під'їзді
	//   - error: nil при успіху, помилки при проблемах з БД
	FindByEntrance(entrance int) ([]*domain.Apartment, error)

	// Update оновлює існуючу квартиру в БД.
	// Автоматично оновлює поле UpdatedAt.
	//
	// Параметри:
	//   - apartment: вказівник на domain.Apartment з оновленими даними
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не існує
	Update(apartment *domain.Apartment) error

	// Deactivate виконує "м'яке" видалення квартири.
	// Встановлює IsActive = false замість фізичного видалення.
	//
	// Параметри:
	//   - id: ідентифікатор квартири
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не існує
	Deactivate(id int) error

	// Activate активує раніше деактивовану квартиру.
	//
	// Параметри:
	//   - id: ідентифікатор квартири
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не існує
	Activate(id int) error

	// Delete виконує фізичне видалення квартири з БД.
	// УВАГА: Використовувати тільки якщо немає пов'язаних платежів!
	//
	// Параметри:
	//   - id: ідентифікатор квартири
	//
	// Повертає:
	//   - error: nil при успіху, domain.ErrApartmentNotFound якщо не існує
	Delete(id int) error

	// Count повертає загальну кількість квартир (активних).
	//
	// Повертає:
	//   - int: кількість квартир
	//   - error: nil при успіху, помилки при проблемах з БД
	Count() (int, error)

	// GetStatistics повертає статистику по квартирах.
	// Корисно для відображення на dashboard.
	//
	// Повертає:
	//   - *ApartmentStatistics: статистичні дані
	//   - error: nil при успіху, помилки при проблемах з БД
	GetStatistics() (*ApartmentStatistics, error)
}

// ApartmentStatistics містить статистичну інформацію про квартири.
type ApartmentStatistics struct {
	// TotalApartments - загальна кількість активних квартир
	TotalApartments int

	// TotalArea - загальна площа всіх квартир
	TotalArea float64

	// AverageArea - середня площа квартири
	AverageArea float64

	// TotalResidents - загальна кількість зареєстрованих мешканців
	TotalResidents int

	// MinFloor - найнижчий поверх
	MinFloor int

	// MaxFloor - найвищий поверх
	MaxFloor int

	// EntrancesCount - кількість під'їздів (якщо використовуються)
	EntrancesCount int
}
