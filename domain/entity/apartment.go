// domain/entity/apartment.go
package entity

import (
	"errors"
	"fmt"
	"time"
)

// Apartment представляє квартиру.
type Apartment struct {
	ID              int64
	ApartmentNumber string
	Floor           int
	Entrance        *int // Під'їзд (опціонально)
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
	IsActive        bool
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validation errors
var (
	ErrApartmentNumberRequired         = errors.New("apartment number is required")
	ErrApartmentNumberTooLong          = errors.New("apartment number must not exceed 20 characters")
	ErrApartmentFloorInvalid           = errors.New("floor must be >= 0")
	ErrApartmentEntranceInvalid        = errors.New("entrance must be > 0")
	ErrApartmentAreaTotalInvalid       = errors.New("total area must be > 0")
	ErrApartmentAreaLivingInvalid      = errors.New("living area must be > 0 and <= total area")
	ErrApartmentRoomsCountInvalid      = errors.New("rooms count must be > 0")
	ErrApartmentCadastralNumberTooLong = errors.New("cadastral number must not exceed 50 characters")
)

// NewApartment створює нову квартиру з валідацією.
func NewApartment(
	apartmentNumber string,
	floor int,
	areaTotal float64,
	entrance *int,
	areaLiving *float64,
	roomsCount *int,
) (*Apartment, error) {
	apt := &Apartment{
		ApartmentNumber: apartmentNumber,
		Floor:           floor,
		AreaTotal:       areaTotal,
		Entrance:        entrance,
		AreaLiving:      areaLiving,
		RoomsCount:      roomsCount,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := apt.Validate(); err != nil {
		return nil, err
	}

	return apt, nil
}

// Validate перевіряє коректність всіх полів.
func (a *Apartment) Validate() error {
	// Apartment number
	if a.ApartmentNumber == "" {
		return ErrApartmentNumberRequired
	}
	if len(a.ApartmentNumber) > 20 {
		return ErrApartmentNumberTooLong
	}

	// Floor
	if a.Floor < 0 {
		return ErrApartmentFloorInvalid
	}

	// Entrance
	if a.Entrance != nil && *a.Entrance <= 0 {
		return ErrApartmentEntranceInvalid
	}

	// Total area
	if a.AreaTotal <= 0 {
		return ErrApartmentAreaTotalInvalid
	}

	// Living area
	if a.AreaLiving != nil {
		if *a.AreaLiving <= 0 || *a.AreaLiving > a.AreaTotal {
			return ErrApartmentAreaLivingInvalid
		}
	}

	// Rooms count
	if a.RoomsCount != nil && *a.RoomsCount <= 0 {
		return ErrApartmentRoomsCountInvalid
	}

	// Cadastral number
	if a.CadastralNumber != nil && len(*a.CadastralNumber) > 50 {
		return ErrApartmentCadastralNumberTooLong
	}

	return nil
}

// Update оновлює дані квартири.
func (a *Apartment) Update(
	floor int,
	areaTotal float64,
	entrance *int,
	areaLiving *float64,
	roomsCount *int,
	cadastralNumber *string,
	notes *string,
) error {
	a.Floor = floor
	a.AreaTotal = areaTotal
	a.Entrance = entrance
	a.AreaLiving = areaLiving
	a.RoomsCount = roomsCount
	a.CadastralNumber = cadastralNumber
	a.Notes = notes
	a.UpdatedAt = time.Now()

	return a.Validate()
}

// Deactivate деактивує квартиру.
func (a *Apartment) Deactivate() {
	a.IsActive = false
	a.UpdatedAt = time.Now()
}

// Activate активує квартиру.
func (a *Apartment) Activate() {
	a.IsActive = true
	a.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення квартири.
func (a *Apartment) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.IsActive = false
	a.UpdatedAt = now
}

// IsDeleted перевіряє чи видалена квартира.
func (a *Apartment) IsDeleted() bool {
	return a.DeletedAt != nil
}

// GetDisplayName повертає відображуване ім'я квартири.
func (a *Apartment) GetDisplayName() string {
	return "Кв. " + a.ApartmentNumber
}

// GetFullInfo повертає повну інформацію про квартиру.
func (a *Apartment) GetFullInfo() string {
	info := a.GetDisplayName()

	if a.Entrance != nil {
		info += ", під'їзд " + string(rune(*a.Entrance+'0'))
	}

	info += ", " + string(rune(a.Floor+'0')) + " пов."

	return info
}

// GetAreaInfo повертає інформацію про площі.
func (a *Apartment) GetAreaInfo() string {
	info := fmt.Sprintf("%.1f м²", a.AreaTotal)

	if a.AreaLiving != nil {
		info += fmt.Sprintf(" (житл. %.1f м²)", *a.AreaLiving)
	}

	return info
}
