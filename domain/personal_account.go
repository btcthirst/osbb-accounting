package domain

import "time"

// =============================================================================
// PersonalAccount (Особистий Рахунок)
// =============================================================================

// PersonalAccount представляє особистий рахунок для обліку платежів.
// Це унікальний фінансовий код для кожної квартири, що використовується
// в бухгалтерії та при прийомі платежів через банк.
type PersonalAccount struct {
	// ID - унікальний ідентифікатор особистого рахунку
	ID int `json:"id" db:"id"`

	// AccountNumber - номер особистого рахунку (унікальний у системі)
	// Формат: XXXX-XXXX або будь-який інший зручний формат
	AccountNumber string `json:"account_number" db:"account_number"`

	// ApartmentID - ID квартири, до якої прив'язаний рахунок
	ApartmentID int `json:"apartment_id" db:"apartment_id"`

	// OwnerID - ID власника рахунку
	OwnerID int `json:"owner_id" db:"owner_id"`

	// OpenedAt - дата відкриття рахунку
	OpenedAt time.Time `json:"opened_at" db:"opened_at"`

	// ClosedAt - дата закриття рахунку (NULL якщо активний)
	ClosedAt *time.Time `json:"closed_at,omitempty" db:"closed_at"`

	// CurrentBalance - поточний баланс рахунку
	// Додатній = переплата, від'ємний = борг
	CurrentBalance float64 `json:"current_balance" db:"current_balance"`

	// IsActive - чи активний рахунок
	IsActive bool `json:"is_active" db:"is_active"`

	// Notes - примітки про рахунок
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
	if pa.ApartmentID <= 0 {
		return ErrInvalidApartmentID
	}
	if pa.OwnerID <= 0 {
		return ErrInvalidOwnerID
	}
	return nil
}

// HasDebt перевіряє, чи є заборгованість на рахунку.
func (pa *PersonalAccount) HasDebt() bool {
	return pa.CurrentBalance < 0
}

// GetDebtAmount повертає суму боргу (завжди додатня).
func (pa *PersonalAccount) GetDebtAmount() float64 {
	if pa.CurrentBalance < 0 {
		return -pa.CurrentBalance
	}
	return 0
}

// GetOverpaymentAmount повертає суму переплати (завжди додатня).
func (pa *PersonalAccount) GetOverpaymentAmount() float64 {
	if pa.CurrentBalance > 0 {
		return pa.CurrentBalance
	}
	return 0
}

// IsClosed перевіряє, чи закритий рахунок.
func (pa *PersonalAccount) IsClosed() bool {
	return pa.ClosedAt != nil
}

// Close закриває особистий рахунок.
func (pa *PersonalAccount) Close() {
	now := time.Now()
	pa.ClosedAt = &now
	pa.IsActive = false
}

// =============================================================================
// Допоміжні структури
// =============================================================================

// OwnerWithAccounts містить інформацію про власника з його рахунками.
type OwnerWithAccounts struct {
	Owner    *Owner
	Accounts []*PersonalAccount
}

// ApartmentWithOwner містить інформацію про квартиру з власником.
type ApartmentWithOwner struct {
	Apartment *Apartment
	Owner     *Owner
	Account   *PersonalAccount
}

// OSBBSettings містить налаштування ОСББ для швидкого доступу.
type OSBBSettings struct {
	OSBB            *OSBB
	TotalApartments int
	TotalArea       float64
	TotalOwners     int
	ActiveAccounts  int
}
