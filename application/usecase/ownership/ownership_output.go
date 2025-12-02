package ownership

import (
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"time"
)

// ============================================================================
// OUTPUT TYPES
// ============================================================================

// OwnershipShareOutput - результат операцій з часткою власності.
type OwnershipShareOutput struct {
	ID                int64
	OwnerID           int64
	ApartmentID       int64
	ShareNumerator    int
	ShareDenominator  int
	ShareFraction     string
	SharePercentage   float64
	OwnershipType     entity.OwnershipType
	OwnershipTypeName string
	StartDate         time.Time
	EndDate           *time.Time
	DocumentType      *string
	DocumentNumber    *string
	DocumentDate      *time.Time
	Notes             *string
	IsActive          bool
	IsCurrentlyActive bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// OwnershipShareDetailsOutput - детальна інформація про частку.
type OwnershipShareDetailsOutput struct {
	OwnershipShareOutput
	OwnerName         string
	OwnerPhone        *string
	OwnerEmail        *string
	ApartmentNumber   string
	ApartmentFloor    int
	ApartmentEntrance *int
	AreaOwned         float64
}

type DeleteOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}
