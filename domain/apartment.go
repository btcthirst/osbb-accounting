package domain

import (
	"fmt"
	"time"
)

// Apartment представляє квартиру в ОСББ.
// Це ключова сутність для обліку платежів та нарахувань.
type Apartment struct {
	// ID - унікальний ідентифікатор квартири
	ID int `json:"id" db:"id"`

	// ApartmentNumber - номер квартири (може містити букви, наприклад "12А")
	ApartmentNumber string `json:"apartment_number" db:"apartment_number"`

	// Floor - поверх, на якому розташована квартира
	Floor int `json:"floor" db:"floor"`

	// Entrance - під'їзд (опціонально, може бути NULL)
	Entrance *int `json:"entrance,omitempty" db:"entrance"`

	// Area - площа квартири в квадратних метрах
	Area float64 `json:"area" db:"area"`

	// Rooms - кількість кімнат
	Rooms int `json:"rooms" db:"rooms"`

	// OwnerName - ПІБ власника квартири
	OwnerName string `json:"owner_name" db:"owner_name"`

	// OwnerPhone - контактний телефон власника
	OwnerPhone string `json:"owner_phone" db:"owner_phone"`

	// OwnerEmail - email власника (опціонально)
	OwnerEmail string `json:"owner_email,omitempty" db:"owner_email"`

	// ResidentsCount - кількість зареєстрованих мешканців
	// Використовується для розрахунку деяких платежів
	ResidentsCount int `json:"residents_count" db:"residents_count"`

	// Notes - додаткові примітки про квартиру
	Notes string `json:"notes,omitempty" db:"notes"`

	// IsActive - чи активна квартира (для м'якого видалення)
	IsActive bool `json:"is_active" db:"is_active"`

	// CreatedAt - час створення запису
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt - час останнього оновлення
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate виконує валідацію полів квартири.
//
// Повертає:
//   - error: опис проблеми або nil якщо все валідне
func (a *Apartment) Validate() error {
	if a.ApartmentNumber == "" {
		return ErrApartmentNumberRequired
	}

	if a.Floor < 0 {
		return ErrInvalidFloor
	}

	if a.Area <= 0 {
		return ErrInvalidArea
	}

	if a.Rooms < 0 {
		return ErrInvalidRooms
	}

	if a.OwnerName == "" {
		return ErrOwnerNameRequired
	}

	if a.ResidentsCount < 0 {
		return ErrInvalidResidentsCount
	}

	return nil
}

// GetDisplayName повертає зручне для відображення ім'я квартири.
//
// Повертає:
//   - string: форматований рядок, наприклад "Кв. 42 (5 поверх)"
func (a *Apartment) GetDisplayName() string {
	if a.Entrance != nil {
		return fmt.Sprintf("Кв. %s (під'їзд %d, поверх %d)",
			a.ApartmentNumber, *a.Entrance, a.Floor)
	}
	return fmt.Sprintf("Кв. %s (поверх %d)", a.ApartmentNumber, a.Floor)
}

// GetOwnerInfo повертає контактну інформацію власника.
//
// Повертає:
//   - string: форматований рядок з ПІБ та телефоном
func (a *Apartment) GetOwnerInfo() string {
	info := a.OwnerName
	if a.OwnerPhone != "" {
		info += " (" + a.OwnerPhone + ")"
	}
	return info
}

// CalculateBaseFee розраховує базовий платіж за площу.
// Використовується для нарахування деяких видів платежів.
//
// Параметри:
//   - ratePerSqm: тариф за квадратний метр
//
// Повертає:
//   - float64: сума до сплати
func (a *Apartment) CalculateBaseFee(ratePerSqm float64) float64 {
	return a.Area * ratePerSqm
}

// CalculateResidentsFee розраховує платіж на основі кількості мешканців.
// Використовується для комунальних послуг, які залежать від кількості осіб.
//
// Параметри:
//   - ratePerPerson: тариф на одну особу
//
// Повертає:
//   - float64: сума до сплати
func (a *Apartment) CalculateResidentsFee(ratePerPerson float64) float64 {
	if a.ResidentsCount <= 0 {
		return 0
	}
	return float64(a.ResidentsCount) * ratePerPerson
}

// ApartmentSummary містить статистичну інформацію про квартиру.
// Використовується для відображення в списках та звітах.
type ApartmentSummary struct {
	Apartment    *Apartment // Основна інформація про квартиру
	TotalDebt    float64    // Загальна заборгованість
	LastPayment  *time.Time // Дата останнього платежу
	PaymentCount int        // Кількість платежів за останній період
}

// ApartmentFilter містить параметри для фільтрації квартир.
// Використовується в Repository для пошуку.
type ApartmentFilter struct {
	// SearchQuery - пошук за номером квартири або іменем власника
	SearchQuery string

	// Floor - фільтр по поверху (nil = всі поверхи)
	Floor *int

	// Entrance - фільтр по під'їзду (nil = всі під'їзди)
	Entrance *int

	// MinArea - мінімальна площа
	MinArea *float64

	// MaxArea - максимальна площа
	MaxArea *float64

	// Rooms - кількість кімнат (nil = будь-яка)
	Rooms *int

	// HasDebt - показувати тільки з боргом (nil = всі)
	HasDebt *bool

	// IsActive - фільтр по активності (nil = всі)
	IsActive *bool

	// SortBy - поле для сортування (apartment_number, floor, area, owner_name)
	SortBy string

	// SortDesc - сортування в зворотному порядку
	SortDesc bool
}

// DefaultApartmentFilter повертає фільтр за замовчуванням.
func DefaultApartmentFilter() *ApartmentFilter {
	isActive := true
	return &ApartmentFilter{
		IsActive: &isActive,
		SortBy:   "apartment_number",
		SortDesc: false,
	}
}
