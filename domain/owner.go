package domain

import "time"

// =============================================================================
// Owner (Власник Квартири)
// =============================================================================

// Owner представляє власника квартири.
// Це фізична особа, яка може володіти однією або кількома квартирами.
type Owner struct {
	// ID - унікальний ідентифікатор власника
	ID int `json:"id" db:"id"`

	// FullName - повне ПІБ власника
	FullName string `json:"full_name" db:"full_name"`

	// TaxID - індивідуальний податковий номер (ІПН/РНОКПП)
	// 10 цифр для фізичних осіб
	TaxID string `json:"tax_id" db:"tax_id"`

	// Phone - основний контактний телефон
	Phone string `json:"phone" db:"phone"`

	// Email - email адреса (опціонально)
	Email string `json:"email,omitempty" db:"email"`

	// AlternativePhone - додатковий телефон (опціонально)
	AlternativePhone string `json:"alternative_phone,omitempty" db:"alternative_phone"`

	// PassportSeries - серія паспорта (опціонально)
	PassportSeries string `json:"passport_series,omitempty" db:"passport_series"`

	// PassportNumber - номер паспорта (опціонально)
	PassportNumber string `json:"passport_number,omitempty" db:"passport_number"`

	// Notes - додаткові примітки про власника
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
		return ErrOwnerNameRequired
	}
	if o.TaxID == "" {
		return ErrTaxIDRequired
	}
	// Валідація ІПН (10 цифр)
	if len(o.TaxID) != 10 {
		return ErrInvalidTaxID
	}
	if o.Phone == "" {
		return ErrPhoneRequired
	}
	return nil
}

// GetContactInfo повертає контактну інформацію власника.
func (o *Owner) GetContactInfo() string {
	info := o.Phone
	if o.Email != "" {
		info += " | " + o.Email
	}
	return info
}
