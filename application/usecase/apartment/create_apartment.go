// application/usecase/apartment/create_apartment.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateApartmentUseCase створює нову квартиру.
type CreateApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateApartmentUseCase створює новий use case.
func NewCreateApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *CreateApartmentUseCase {
	return &CreateApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateApartmentInput - вхідні дані.
type CreateApartmentInput struct {
	CurrentUserID   int64
	ApartmentNumber string
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
}

// Execute виконує створення квартири.
func (uc *CreateApartmentUseCase) Execute(ctx context.Context, input CreateApartmentInput) (*ApartmentOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Перевірка унікальності номеру
	exists, err := uc.apartmentRepo.ExistsByNumber(ctx, input.ApartmentNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check apartment number: %w", err)
	}
	if exists {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"apartment with this number already exists",
			domainErrors.ErrAlreadyExists,
		).WithDetails("field", "apartment_number")
	}

	// Створення entity
	apartment, err := entity.NewApartment(
		input.ApartmentNumber,
		input.Floor,
		input.AreaTotal,
		input.Entrance,
		input.AreaLiving,
		input.RoomsCount,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid apartment data",
			err,
		)
	}

	apartment.CadastralNumber = input.CadastralNumber
	apartment.Notes = input.Notes

	// Збереження
	if err := uc.apartmentRepo.Create(ctx, apartment); err != nil {
		return nil, fmt.Errorf("failed to create apartment: %w", err)
	}

	return mapApartmentToOutput(apartment), nil
}
