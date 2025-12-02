// application/usecase/apartment/update_apartment.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateApartmentUseCase оновлює дані квартири.
type UpdateApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateApartmentUseCase створює новий use case.
func NewUpdateApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateApartmentUseCase {
	return &UpdateApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateApartmentInput - вхідні дані.
type UpdateApartmentInput struct {
	CurrentUserID   int64
	ApartmentID     int64
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
}

// Execute виконує оновлення квартири.
func (uc *UpdateApartmentUseCase) Execute(ctx context.Context, input UpdateApartmentInput) (*ApartmentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	apartment, err := uc.apartmentRepo.GetByID(ctx, input.ApartmentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"apartment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	if err := apartment.Update(
		input.Floor,
		input.AreaTotal,
		input.Entrance,
		input.AreaLiving,
		input.RoomsCount,
		input.CadastralNumber,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid apartment data",
			err,
		)
	}

	if err := uc.apartmentRepo.Update(ctx, apartment); err != nil {
		return nil, fmt.Errorf("failed to update apartment: %w", err)
	}

	return mapApartmentToOutput(apartment), nil
}
