package domain

import "time"

// OSBB представляє організацію ОСББ (єдина на систему).
type OSBB struct {
	// ID - завжди 1 (єдина організація)
	ID int `json:"id" db:"id"`

	// Name - повна назва ОСББ
	Name string `json:"name" db:"name"`

	// ShortName - скорочена назва
	ShortName string `json:"short_name" db:"short_name"`

	// Address - повна адреса будинку
	Address string `json:"address" db:"address"`

	// EDRPOU - код ЄДРПОУ організації
	EDRPOU string `json:"edrpou" db:"edrpou"`

	// HeadOfBoard - ПІБ голови правління
	HeadOfBoard string `json:"head_of_board" db:"head_of_board"`

	// BaseMaintenanceRate - базовий тариф на утримання (грн/м²)
	BaseMaintenanceRate float64 `json:"base_maintenance_rate" db:"base_maintenance_rate"`

	// BaseUtilitiesRate - базовий тариф на комунальні (грн/м²)
	BaseUtilitiesRate float64 `json:"base_utilities_rate" db:"base_utilities_rate"`

	// BankName - назва банку
	BankName string `json:"bank_name" db:"bank_name"`

	// BankAccount - розрахунковий рахунок ОСББ
	BankAccount string `json:"bank_account" db:"bank_account"`

	// BankMFO - МФО банку
	BankMFO string `json:"bank_mfo" db:"bank_mfo"`

	// Phone - контактний телефон
	Phone string `json:"phone" db:"phone"`

	// Email - контактний email
	Email string `json:"email" db:"email"`

	// CreatedAt - час створення
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
		return ErrAddressRequired
	}
	if o.BaseMaintenanceRate < 0 {
		return ErrInvalidRate
	}
	if o.BaseUtilitiesRate < 0 {
		return ErrInvalidRate
	}
	return nil
}

// GetFullName повертає повну назву для документів.
func (o *OSBB) GetFullName() string {
	if o.ShortName != "" {
		return o.ShortName
	}
	return o.Name
}
