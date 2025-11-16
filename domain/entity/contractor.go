// domain/entity/contractor.go
package entity

import (
	"errors"
	"regexp"
	"time"
)

// Contractor представляє контрагента (постачальника послуг, підрядника).
type Contractor struct {
	ID             int64
	Name           string
	EDRPOU         *string // ЄДРПОУ (8-10 цифр)
	ContractorType ContractorType
	ContactPerson  *string
	Phone          *string
	Email          *string
	Address        *string
	BankAccount    *string // IBAN (29 символів)
	BankName       *string
	BankMFO        *string // МФО банку (6 цифр)
	ContractNumber *string
	ContractDate   *time.Time
	Notes          *string
	IsActive       bool
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ContractorType представляє тип контрагента.
type ContractorType string

const (
	ContractorTypeUtility  ContractorType = "utility"  // Комунальні послуги
	ContractorTypeService  ContractorType = "service"  // Послуги
	ContractorTypeSupplier ContractorType = "supplier" // Постачальники
	ContractorTypeOther    ContractorType = "other"    // Інше
)

// Validation errors
var (
	ErrContractorNameRequired        = errors.New("contractor name is required")
	ErrContractorNameTooShort        = errors.New("contractor name must be at least 2 characters")
	ErrContractorNameTooLong         = errors.New("contractor name must not exceed 200 characters")
	ErrContractorTypeInvalid         = errors.New("invalid contractor type")
	ErrContractorEDRPOUInvalidFormat = errors.New("EDRPOU must be 8-10 digits")
	ErrContractorPhoneInvalidFormat  = errors.New("phone must be at least 10 digits")
	ErrContractorEmailInvalidFormat  = errors.New("invalid email format")
	ErrContractorBankAccountInvalid  = errors.New("bank account must be exactly 29 characters (IBAN)")
	ErrContractorBankMFOInvalid      = errors.New("bank MFO must be exactly 6 digits")
	ErrContractorContractDateInvalid = errors.New("contract date cannot be in the future")
)

// Regular expressions
var (
	contractorEDRPOURegex  = regexp.MustCompile(`^\d{8,10}$`)
	contractorPhoneRegex   = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	contractorEmailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	contractorBankMFORegex = regexp.MustCompile(`^\d{6}$`)
)

// NewContractor створює нового контрагента з валідацією.
func NewContractor(
	name string,
	contractorType ContractorType,
	edrpou *string,
	contactPerson *string,
	phone *string,
	email *string,
) (*Contractor, error) {
	contractor := &Contractor{
		Name:           name,
		ContractorType: contractorType,
		EDRPOU:         edrpou,
		ContactPerson:  contactPerson,
		Phone:          phone,
		Email:          email,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := contractor.Validate(); err != nil {
		return nil, err
	}

	return contractor, nil
}

// Validate перевіряє коректність всіх полів контрагента.
func (c *Contractor) Validate() error {
	// Name validation
	if c.Name == "" {
		return ErrContractorNameRequired
	}
	if len(c.Name) < 2 {
		return ErrContractorNameTooShort
	}
	if len(c.Name) > 200 {
		return ErrContractorNameTooLong
	}

	// ContractorType validation
	if !c.ContractorType.IsValid() {
		return ErrContractorTypeInvalid
	}

	// EDRPOU validation (optional)
	if c.EDRPOU != nil && *c.EDRPOU != "" {
		if !contractorEDRPOURegex.MatchString(*c.EDRPOU) {
			return ErrContractorEDRPOUInvalidFormat
		}
	}

	// Phone validation (optional)
	if c.Phone != nil && *c.Phone != "" {
		if !contractorPhoneRegex.MatchString(*c.Phone) {
			return ErrContractorPhoneInvalidFormat
		}
	}

	// Email validation (optional)
	if c.Email != nil && *c.Email != "" {
		if !contractorEmailRegex.MatchString(*c.Email) {
			return ErrContractorEmailInvalidFormat
		}
	}

	// Bank account validation (optional, IBAN format - 29 characters)
	if c.BankAccount != nil && *c.BankAccount != "" {
		if len(*c.BankAccount) != 29 {
			return ErrContractorBankAccountInvalid
		}
	}

	// Bank MFO validation (optional, 6 digits)
	if c.BankMFO != nil && *c.BankMFO != "" {
		if !contractorBankMFORegex.MatchString(*c.BankMFO) {
			return ErrContractorBankMFOInvalid
		}
	}

	// Contract date validation (optional, cannot be in the future)
	if c.ContractDate != nil {
		if c.ContractDate.After(time.Now()) {
			return ErrContractorContractDateInvalid
		}
	}

	return nil
}

// IsValid перевіряє чи валідний тип контрагента.
func (ct ContractorType) IsValid() bool {
	switch ct {
	case ContractorTypeUtility, ContractorTypeService, ContractorTypeSupplier, ContractorTypeOther:
		return true
	default:
		return false
	}
}

// Update оновлює дані контрагента.
func (c *Contractor) Update(
	name string,
	contractorType ContractorType,
	contactPerson *string,
	phone *string,
	email *string,
	address *string,
	notes *string,
) error {
	c.Name = name
	c.ContractorType = contractorType
	c.ContactPerson = contactPerson
	c.Phone = phone
	c.Email = email
	c.Address = address
	c.Notes = notes
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// UpdateBankDetails оновлює банківські реквізити.
func (c *Contractor) UpdateBankDetails(
	bankAccount *string,
	bankName *string,
	bankMFO *string,
) error {
	c.BankAccount = bankAccount
	c.BankName = bankName
	c.BankMFO = bankMFO
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// UpdateContract оновлює договірну інформацію.
func (c *Contractor) UpdateContract(
	contractNumber *string,
	contractDate *time.Time,
) error {
	c.ContractNumber = contractNumber
	c.ContractDate = contractDate
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// Deactivate деактивує контрагента.
func (c *Contractor) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// Activate активує контрагента.
func (c *Contractor) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення контрагента.
func (c *Contractor) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.IsActive = false
	c.UpdatedAt = now
}

// IsDeleted перевіряє чи видалений контрагент.
func (c *Contractor) IsDeleted() bool {
	return c.DeletedAt != nil
}

// GetDisplayName повертає відображуване ім'я контрагента.
func (c *Contractor) GetDisplayName() string {
	if c.EDRPOU != nil && *c.EDRPOU != "" {
		return c.Name + " (ЄДРПОУ: " + *c.EDRPOU + ")"
	}
	return c.Name
}

// GetContactInfo повертає контактну інформацію.
func (c *Contractor) GetContactInfo() string {
	contacts := ""
	if c.ContactPerson != nil && *c.ContactPerson != "" {
		contacts += "Контактна особа: " + *c.ContactPerson
	}
	if c.Phone != nil && *c.Phone != "" {
		if contacts != "" {
			contacts += ", "
		}
		contacts += "Тел: " + *c.Phone
	}
	if c.Email != nil && *c.Email != "" {
		if contacts != "" {
			contacts += ", "
		}
		contacts += "Email: " + *c.Email
	}
	return contacts
}

// HasBankDetails перевіряє чи вказані банківські реквізити.
func (c *Contractor) HasBankDetails() bool {
	return (c.BankAccount != nil && *c.BankAccount != "") ||
		(c.BankName != nil && *c.BankName != "") ||
		(c.BankMFO != nil && *c.BankMFO != "")
}

// HasContract перевіряє чи вказані договірні дані.
func (c *Contractor) HasContract() bool {
	return (c.ContractNumber != nil && *c.ContractNumber != "") ||
		c.ContractDate != nil
}

// GetTypeName повертає локалізовану назву типу контрагента.
func (ct ContractorType) GetDisplayName() string {
	switch ct {
	case ContractorTypeUtility:
		return "Комунальні послуги"
	case ContractorTypeService:
		return "Послуги"
	case ContractorTypeSupplier:
		return "Постачальник"
	case ContractorTypeOther:
		return "Інше"
	default:
		return "Невідомо"
	}
}
