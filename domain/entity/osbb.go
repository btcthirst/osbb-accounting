// domain/entity/osbb.go
package entity

import (
	"errors"
	"regexp"
	"time"
	"unicode/utf8"
)

// OSBB представляє організацію ОСББ (Singleton).
// В системі може бути тільки одна організація ОСББ.
type OSBB struct {
	ID            int64
	Name          string
	EDRPOU        string // ЄДРПОУ (8 цифр)
	LegalAddress  string
	ActualAddress *string
	Phone         *string
	Email         *string
	Website       *string
	ChairmanName  string // ПІБ голови правління
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validation errors
var (
	ErrOSBBNameRequired         = errors.New("OSBB name is required")
	ErrOSBBNameTooShort         = errors.New("OSBB name must be at least 3 characters")
	ErrOSBBEDRPOURequired       = errors.New("EDRPOU is required")
	ErrOSBBEDRPOUInvalidFormat  = errors.New("EDRPOU must be exactly 8 digits")
	ErrOSBBLegalAddressRequired = errors.New("legal address is required")
	ErrOSBBChairmanNameRequired = errors.New("chairman name is required")
	ErrOSBBPhoneInvalidFormat   = errors.New("phone must be at least 10 digits")
	ErrOSBBEmailInvalidFormat   = errors.New("invalid email format")
	ErrOSBBWebsiteInvalidFormat = errors.New("invalid website format")
)

// Regular expressions
var (
	edrpouRegex = regexp.MustCompile(`^\d{8}$`)
	//phoneRegex   = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	//emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	websiteRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

// NewOSBB створює нову організацію ОСББ з валідацією.
func NewOSBB(
	name, edrpou, legalAddress, chairmanName string,
	actualAddress, phone, email, website *string,
) (*OSBB, error) {
	osbb := &OSBB{
		Name:          name,
		EDRPOU:        edrpou,
		LegalAddress:  legalAddress,
		ActualAddress: actualAddress,
		Phone:         phone,
		Email:         email,
		Website:       website,
		ChairmanName:  chairmanName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := osbb.Validate(); err != nil {
		return nil, err
	}

	return osbb, nil
}

// Validate перевіряє коректність всіх полів.
func (o *OSBB) Validate() error {
	// Name validation
	if o.Name == "" {
		return ErrOSBBNameRequired
	}
	if utf8.RuneCountInString(o.Name) < 3 {
		return ErrOSBBNameTooShort
	}

	// EDRPOU validation (8 цифр)
	if o.EDRPOU == "" {
		return ErrOSBBEDRPOURequired
	}
	if !edrpouRegex.MatchString(o.EDRPOU) {
		return ErrOSBBEDRPOUInvalidFormat
	}

	// Legal address validation
	if o.LegalAddress == "" {
		return ErrOSBBLegalAddressRequired
	}

	// Chairman name validation
	if o.ChairmanName == "" {
		return ErrOSBBChairmanNameRequired
	}

	// Phone validation (optional)
	if o.Phone != nil && *o.Phone != "" {
		if !phoneRegex.MatchString(*o.Phone) {
			return ErrOSBBPhoneInvalidFormat
		}
	}

	// Email validation (optional)
	if o.Email != nil && *o.Email != "" {
		if !emailRegex.MatchString(*o.Email) {
			return ErrOSBBEmailInvalidFormat
		}
	}

	// Website validation (optional)
	if o.Website != nil && *o.Website != "" {
		if !websiteRegex.MatchString(*o.Website) {
			return ErrOSBBWebsiteInvalidFormat
		}
	}

	return nil
}

// Update оновлює дані ОСББ.
func (o *OSBB) Update(
	name, legalAddress, chairmanName string,
	actualAddress, phone, email, website *string,
) error {
	o.Name = name
	o.LegalAddress = legalAddress
	o.ChairmanName = chairmanName
	o.ActualAddress = actualAddress
	o.Phone = phone
	o.Email = email
	o.Website = website
	o.UpdatedAt = time.Now()

	return o.Validate()
}

// UpdateChairman оновлює ПІБ голови правління.
func (o *OSBB) UpdateChairman(chairmanName string) error {
	if chairmanName == "" {
		return ErrOSBBChairmanNameRequired
	}
	o.ChairmanName = chairmanName
	o.UpdatedAt = time.Now()
	return nil
}

// UpdateContacts оновлює контактні дані.
func (o *OSBB) UpdateContacts(phone, email, website *string) error {
	o.Phone = phone
	o.Email = email
	o.Website = website
	o.UpdatedAt = time.Now()
	return o.Validate()
}

// GetFullAddress повертає повну адресу (юридичну або фактичну).
func (o *OSBB) GetFullAddress() string {
	if o.ActualAddress != nil && *o.ActualAddress != "" {
		return *o.ActualAddress
	}
	return o.LegalAddress
}

// HasContact перевіряє чи вказані контактні дані.
func (o *OSBB) HasContact() bool {
	return (o.Phone != nil && *o.Phone != "") ||
		(o.Email != nil && *o.Email != "") ||
		(o.Website != nil && *o.Website != "")
}
