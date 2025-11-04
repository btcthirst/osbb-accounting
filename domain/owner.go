package domain

import "time"

// Owner представляє власника квартири (фізична особа).
// Один власник може мати декілька квартир.
type Owner struct {
	// ID - унікальний ідентифікатор власника
	ID int `json:"id" db:"id"`

	// FullName - повне ПІБ власника
	FullName string `json:"full_name" db:"full_name"`

	// TaxID - ІПН (індивідуальний податковий номер)
	TaxID string `json:"tax_id" db:"tax_id"`

	// Phone - основний контактний телефон
	Phone string `json:"phone" db:"phone"`

	// PhoneAdditional - додатковий телефон (опціонально)
	PhoneAdditional string `json:"phone_additional,omitempty" db:"phone_additional"`

	// Email - контактний email
	Email string `json:"email,omitempty" db:"email"`

	// PassportSeries - серія паспорту
	PassportSeries string `json:"passport_series,omitempty" db:"passport_series"`

	// PassportNumber - номер паспорту
	PassportNumber string `json:"passport_number,omitempty" db:"passport_number"`

	// PassportIssuedBy - ким виданий паспорт
	PassportIssuedBy string `json:"passport_issued_by,omitempty" db:"passport_issued_by"`

	// PassportIssuedDate - дата видачі паспорту
	PassportIssuedDate *time.Time `json:"passport_issued_date,omitempty" db:"passport_issued_date"`

	// RegistrationAddress - адреса реєстрації
	RegistrationAddress string `json:"registration_address,omitempty" db:"registration_address"`

	// Notes - додаткові примітки
	Notes string `json:"notes,omitempty" db:"notes"`

	// IsActive - чи активний власник
	IsActive bool `json:"is_active" db:"is_active"`

	// CreatedAt - час створення запису
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt - час останнього оновлення
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate виконує валідацію даних власника.
func (o *Owner) Validate() error {
	if o.FullName == "" {
		return ErrFullNameRequired
	}
	if o.TaxID == "" {
		return ErrTaxIDRequired
	}
	if len(o.TaxID) != 10 {
		return ErrInvalidTaxID
	}
	if o.Phone == "" {
		return ErrPhoneRequired
	}
	return nil
}

// GetDisplayName повертає ім'я для відображення.
func (o *Owner) GetDisplayName() string {
	return o.FullName
}

// GetContactInfo повертає контактну інформацію.
func (o *Owner) GetContactInfo() string {
	info := o.Phone
	if o.Email != "" {
		info += " • " + o.Email
	}
	return info
}
