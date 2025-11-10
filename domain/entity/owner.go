// domain/entity/owner.go
package entity

import (
	"errors"
	"regexp"
	"time"
	"unicode/utf8"
)

// Owner представляє власника квартири.
type Owner struct {
	ID                int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string // ІПН (10 цифр)
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
	IsActive          bool
	DeletedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Validation errors
var (
	ErrOwnerFirstNameRequired      = errors.New("first name is required")
	ErrOwnerLastNameRequired       = errors.New("last name is required")
	ErrOwnerNameTooLong            = errors.New("name must not exceed 100 characters")
	ErrOwnerPhoneInvalidFormat     = errors.New("phone must be at least 10 digits")
	ErrOwnerEmailInvalidFormat     = errors.New("invalid email format")
	ErrOwnerTaxNumberInvalidFormat = errors.New("tax number must be exactly 10 digits")
)

// Regular expressions
var (
	ownerPhoneRegex     = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	ownerEmailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	ownerTaxNumberRegex = regexp.MustCompile(`^\d{10}$`)
)

// NewOwner створює нового власника з валідацією.
func NewOwner(
	firstName, lastName string,
	middleName, phone, email, taxNumber *string,
) (*Owner, error) {
	owner := &Owner{
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		Phone:      phone,
		Email:      email,
		TaxNumber:  taxNumber,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := owner.Validate(); err != nil {
		return nil, err
	}

	return owner, nil
}

// Validate перевіряє коректність всіх полів.
func (o *Owner) Validate() error {
	// FirstName validation
	if o.FirstName == "" {
		return ErrOwnerFirstNameRequired
	}
	if utf8.RuneCountInString(o.FirstName) > 100 {
		return ErrOwnerNameTooLong
	}

	// LastName validation
	if o.LastName == "" {
		return ErrOwnerLastNameRequired
	}
	if utf8.RuneCountInString(o.LastName) > 100 {
		return ErrOwnerNameTooLong
	}

	// MiddleName validation (optional)
	if o.MiddleName != nil && utf8.RuneCountInString(*o.MiddleName) > 100 {
		return ErrOwnerNameTooLong
	}

	// Phone validation (optional)
	if o.Phone != nil && *o.Phone != "" {
		if !ownerPhoneRegex.MatchString(*o.Phone) {
			return ErrOwnerPhoneInvalidFormat
		}
	}

	// Email validation (optional)
	if o.Email != nil && *o.Email != "" {
		if !ownerEmailRegex.MatchString(*o.Email) {
			return ErrOwnerEmailInvalidFormat
		}
	}

	// TaxNumber validation (optional, 10 digits)
	if o.TaxNumber != nil && *o.TaxNumber != "" {
		if !ownerTaxNumberRegex.MatchString(*o.TaxNumber) {
			return ErrOwnerTaxNumberInvalidFormat
		}
	}

	return nil
}

// Update оновлює дані власника.
func (o *Owner) Update(
	firstName, lastName string,
	middleName, phone, email *string,
) error {
	o.FirstName = firstName
	o.LastName = lastName
	o.MiddleName = middleName
	o.Phone = phone
	o.Email = email
	o.UpdatedAt = time.Now()

	return o.Validate()
}

// UpdateTaxNumber оновлює ІПН.
func (o *Owner) UpdateTaxNumber(taxNumber *string) error {
	o.TaxNumber = taxNumber
	o.UpdatedAt = time.Now()
	return o.Validate()
}

// UpdatePassport оновлює паспортні дані.
func (o *Owner) UpdatePassport(series, number *string) {
	o.PassportSeries = series
	o.PassportNumber = number
	o.UpdatedAt = time.Now()
}

// UpdateAddresses оновлює адреси.
func (o *Owner) UpdateAddresses(registered, actual *string) {
	o.RegisteredAddress = registered
	o.ActualAddress = actual
	o.UpdatedAt = time.Now()
}

// Deactivate деактивує власника.
func (o *Owner) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

// Activate активує власника.
func (o *Owner) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення власника.
func (o *Owner) SoftDelete() {
	now := time.Now()
	o.DeletedAt = &now
	o.IsActive = false
	o.UpdatedAt = now
}

// IsDeleted перевіряє чи видалений власник.
func (o *Owner) IsDeleted() bool {
	return o.DeletedAt != nil
}

// FullName повертає повне ім'я власника.
func (o *Owner) FullName() string {
	if o.MiddleName != nil && *o.MiddleName != "" {
		return o.LastName + " " + o.FirstName + " " + *o.MiddleName
	}
	return o.LastName + " " + o.FirstName
}

// ShortName повертає коротке ім'я (Прізвище І.П.)
func (o *Owner) ShortName() string {
	result := o.LastName + " " + string([]rune(o.FirstName)[0]) + "."
	if o.MiddleName != nil && *o.MiddleName != "" {
		result += string([]rune(*o.MiddleName)[0]) + "."
	}
	return result
}

// GetContactInfo повертає контактну інформацію.
func (o *Owner) GetContactInfo() string {
	contacts := ""
	if o.Phone != nil && *o.Phone != "" {
		contacts += "Тел: " + *o.Phone
	}
	if o.Email != nil && *o.Email != "" {
		if contacts != "" {
			contacts += ", "
		}
		contacts += "Email: " + *o.Email
	}
	return contacts
}

// HasContact перевіряє чи вказані контактні дані.
func (o *Owner) HasContact() bool {
	return (o.Phone != nil && *o.Phone != "") || (o.Email != nil && *o.Email != "")
}

// HasPassport перевіряє чи вказані паспортні дані.
func (o *Owner) HasPassport() bool {
	return (o.PassportSeries != nil && *o.PassportSeries != "") &&
		(o.PassportNumber != nil && *o.PassportNumber != "")
}
