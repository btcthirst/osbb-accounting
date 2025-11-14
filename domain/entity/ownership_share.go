// domain/entity/ownership_share.go
package entity

import (
	"errors"
	"fmt"
	"time"
)

// OwnershipShare представляє частку власності (зв'язок Owner ↔ Apartment).
type OwnershipShare struct {
	ID               int64
	OwnerID          int64
	ApartmentID      int64
	ShareNumerator   int // Чисельник дробу (напр. 1 з 1/2)
	ShareDenominator int // Знаменник дробу (напр. 2 з 1/2)
	OwnershipType    OwnershipType
	StartDate        time.Time
	EndDate          *time.Time
	DocumentType     *string
	DocumentNumber   *string
	DocumentDate     *time.Time
	Notes            *string
	IsActive         bool
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// OwnershipType представляє тип власності.
type OwnershipType string

const (
	OwnershipTypeFull   OwnershipType = "full"   // Повна власність
	OwnershipTypeShared OwnershipType = "shared" // Часткова власність
	OwnershipTypeRent   OwnershipType = "rent"   // Оренда
)

// Validation errors
var (
	ErrOwnershipOwnerIDRequired         = errors.New("owner ID is required")
	ErrOwnershipApartmentIDRequired     = errors.New("apartment ID is required")
	ErrOwnershipShareNumeratorInvalid   = errors.New("share numerator must be > 0")
	ErrOwnershipShareDenominatorInvalid = errors.New("share denominator must be >= numerator")
	ErrOwnershipTypeInvalid             = errors.New("invalid ownership type")
	ErrOwnershipEndDateInvalid          = errors.New("end date must be after start date")
	ErrOwnershipShareExceedsTotal       = errors.New("total ownership shares exceed 100%")
)

// NewOwnershipShare створює нову частку власності з валідацією.
func NewOwnershipShare(
	ownerID, apartmentID int64,
	shareNumerator, shareDenominator int,
	ownershipType OwnershipType,
	startDate time.Time,
) (*OwnershipShare, error) {
	share := &OwnershipShare{
		OwnerID:          ownerID,
		ApartmentID:      apartmentID,
		ShareNumerator:   shareNumerator,
		ShareDenominator: shareDenominator,
		OwnershipType:    ownershipType,
		StartDate:        startDate,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := share.Validate(); err != nil {
		return nil, err
	}

	return share, nil
}

// Validate перевіряє коректність всіх полів.
func (os *OwnershipShare) Validate() error {
	if os.OwnerID <= 0 {
		return ErrOwnershipOwnerIDRequired
	}

	if os.ApartmentID <= 0 {
		return ErrOwnershipApartmentIDRequired
	}

	if os.ShareNumerator <= 0 {
		return ErrOwnershipShareNumeratorInvalid
	}

	if os.ShareDenominator < os.ShareNumerator {
		return ErrOwnershipShareDenominatorInvalid
	}

	if !os.OwnershipType.IsValid() {
		return ErrOwnershipTypeInvalid
	}

	if os.EndDate != nil && os.EndDate.Before(os.StartDate) {
		return ErrOwnershipEndDateInvalid
	}

	return nil
}

// IsValid перевіряє чи валідний тип власності.
func (ot OwnershipType) IsValid() bool {
	switch ot {
	case OwnershipTypeFull, OwnershipTypeShared, OwnershipTypeRent:
		return true
	default:
		return false
	}
}

// Update оновлює частку власності.
func (os *OwnershipShare) Update(
	shareNumerator, shareDenominator int,
	ownershipType OwnershipType,
	endDate *time.Time,
	documentType, documentNumber *string,
	documentDate *time.Time,
	notes *string,
) error {
	os.ShareNumerator = shareNumerator
	os.ShareDenominator = shareDenominator
	os.OwnershipType = ownershipType
	os.EndDate = endDate
	os.DocumentType = documentType
	os.DocumentNumber = documentNumber
	os.DocumentDate = documentDate
	os.Notes = notes
	os.UpdatedAt = time.Now()

	return os.Validate()
}

// SetDocument встановлює документальне підтвердження.
func (os *OwnershipShare) SetDocument(
	documentType, documentNumber string,
	documentDate time.Time,
) {
	os.DocumentType = &documentType
	os.DocumentNumber = &documentNumber
	os.DocumentDate = &documentDate
	os.UpdatedAt = time.Now()
}

// TerminateOwnership завершує власність.
func (os *OwnershipShare) TerminateOwnership(endDate time.Time) error {
	if endDate.Before(os.StartDate) {
		return ErrOwnershipEndDateInvalid
	}
	os.EndDate = &endDate
	os.IsActive = false
	os.UpdatedAt = time.Now()
	return nil
}

// SoftDelete виконує м'яке видалення.
func (os *OwnershipShare) SoftDelete() {
	now := time.Now()
	os.DeletedAt = &now
	os.IsActive = false
	os.UpdatedAt = now
}

// IsDeleted перевіряє чи видалена частка.
func (os *OwnershipShare) IsDeleted() bool {
	return os.DeletedAt != nil
}

// GetSharePercentage повертає відсоток власності.
func (os *OwnershipShare) GetSharePercentage() float64 {
	if os.ShareDenominator == 0 {
		return 0
	}
	return float64(os.ShareNumerator) / float64(os.ShareDenominator) * 100
}

// GetShareFraction повертає частку у вигляді дробу.
func (os *OwnershipShare) GetShareFraction() string {
	return fmt.Sprintf("%d/%d", os.ShareNumerator, os.ShareDenominator)
}

// GetShareDisplay повертає відображення частки з відсотками.
func (os *OwnershipShare) GetShareDisplay() string {
	return fmt.Sprintf("%s (%.1f%%)",
		os.GetShareFraction(),
		os.GetSharePercentage())
}

// IsCurrentlyActive перевіряє чи активна зараз частка.
func (os *OwnershipShare) IsCurrentlyActive() bool {
	now := time.Now()

	if !os.IsActive || os.DeletedAt != nil {
		return false
	}

	// Перевірка періоду
	if now.Before(os.StartDate) {
		return false
	}

	if os.EndDate != nil && now.After(*os.EndDate) {
		return false
	}

	return true
}

// GetOwnershipTypeName повертає назву типу власності.
func (ot OwnershipType) GetDisplayName() string {
	switch ot {
	case OwnershipTypeFull:
		return "Повна власність"
	case OwnershipTypeShared:
		return "Часткова власність"
	case OwnershipTypeRent:
		return "Оренда"
	default:
		return "Невідомо"
	}
}

// SimplifyFraction спрощує дріб (знаходить НСД).
func SimplifyFraction(numerator, denominator int) (int, int) {
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	divisor := gcd(numerator, denominator)
	return numerator / divisor, denominator / divisor
}
