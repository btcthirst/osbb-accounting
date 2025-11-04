package domain

import "time"

// PersonalAccount представляє особистий рахунок (зв'язок власника та квартири).
// Це ключовий елемент для бухгалтерського та банківського обліку.
type PersonalAccount struct {
	// ID - унікальний ідентифікатор особистого рахунку
	ID int `json:"id" db:"id"`

	// AccountNumber - унікальний номер особистого рахунку
	// Формат: зазвичай 10-12 цифр (наприклад: 1234567890)
	AccountNumber string `json:"account_number" db:"account_number"`

	// ApartmentID - ID квартири
	ApartmentID int `json:"apartment_id" db:"apartment_id"`

	// OwnerID - ID власника
	OwnerID int `json:"owner_id" db:"owner_id"`

	// OpenDate - дата відкриття рахунку
	OpenDate time.Time `json:"open_date" db:"open_date"`

	// CloseDate - дата закриття рахунку (якщо закритий)
	CloseDate *time.Time `json:"close_date,omitempty" db:"close_date"`

	// IsActive - чи активний рахунок
	IsActive bool `json:"is_active" db:"is_active"`

	// Notes - додаткові примітки
	Notes string `json:"notes,omitempty" db:"notes"`

	// CreatedAt - час створення запису
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt - час останнього оновлення
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate виконує валідацію особистого рахунку.
func (pa *PersonalAccount) Validate() error {
	if pa.AccountNumber == "" {
		return ErrAccountNumberRequired
	}
	if len(pa.AccountNumber) < 8 || len(pa.AccountNumber) > 12 {
		return ErrInvalidAccountNumber
	}
	if pa.ApartmentID <= 0 {
		return ErrInvalidApartmentID
	}
	if pa.OwnerID <= 0 {
		return ErrOwnerIDRequired
	}
	return nil
}

// IsClosed перевіряє, чи закритий рахунок.
func (pa *PersonalAccount) IsClosed() bool {
	return pa.CloseDate != nil
}

// PersonalAccountWithDetails містить повну інформацію про особистий рахунок.
type PersonalAccountWithDetails struct {
	Account   *PersonalAccount // Основна інформація
	Apartment *Apartment       // Дані квартири
	Owner     *Owner           // Дані власника
	Balance   float64          // Поточний баланс
	Summary   *PaymentSummary  // Зведення по платежах
}
