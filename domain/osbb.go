package domain

import "time"

// =============================================================================
// OSBB (Об'єднання Співвласників Багатоквартирного Будинку)
// =============================================================================

// OSBB представляє організацію ОСББ.
// Це центральна сутність системи - єдина організація, яка управляє будинком.
type OSBB struct {
	// ID - унікальний ідентифікатор ОСББ
	ID int `json:"id" db:"id"`

	// Name - повна назва ОСББ
	Name string `json:"name" db:"name"`

	// ShortName - скорочена назва (опціонально)
	ShortName string `json:"short_name,omitempty" db:"short_name"`

	// Address - повна адреса будинку
	Address string `json:"address" db:"address"`

	// EDRPOU - код ЄДРПОУ (ідентифікаційний код юридичної особи)
	EDRPOU string `json:"edrpou" db:"edrpou"`

	// BaseRate - базовий тариф за утримання будинку (грн/м²)
	BaseRate float64 `json:"base_rate" db:"base_rate"`

	// ChairmanName - ПІБ голови правління
	ChairmanName string `json:"chairman_name" db:"chairman_name"`

	// ChairmanPhone - контактний телефон голови
	ChairmanPhone string `json:"chairman_phone" db:"chairman_phone"`

	// ChairmanEmail - email голови (опціонально)
	ChairmanEmail string `json:"chairman_email,omitempty" db:"chairman_email"`

	// BankName - назва банку для розрахункового рахунку
	BankName string `json:"bank_name" db:"bank_name"`

	// BankAccount - розрахунковий рахунок ОСББ (IBAN)
	BankAccount string `json:"bank_account" db:"bank_account"`

	// MFO - МФО банку
	MFO string `json:"mfo" db:"mfo"`

	// FoundedAt - дата створення ОСББ
	FoundedAt time.Time `json:"founded_at" db:"founded_at"`

	// CreatedAt - час створення запису в системі
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt - час останнього оновлення
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate виконує валідацію даних ОСББ.
func (o *OSBB) Validate() error {
	if o.Name == "" {
		return ErrOSBBNameRequired
	}
	if o.Address == "" {
		return ErrOSBBAddressRequired
	}
	if o.EDRPOU == "" || len(o.EDRPOU) != 8 {
		return ErrInvalidEDRPOU
	}
	if o.BaseRate < 0 {
		return ErrInvalidBaseRate
	}
	if o.ChairmanName == "" {
		return ErrChairmanNameRequired
	}
	if o.BankAccount == "" {
		return ErrBankAccountRequired
	}
	return nil
}

// GetDisplayName повертає зручне ім'я для відображення.
func (o *OSBB) GetDisplayName() string {
	if o.ShortName != "" {
		return o.ShortName
	}
	return o.Name
}
