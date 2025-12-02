// application/usecase/apartment/get_apartment.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetApartmentUseCase отримує квартиру за ID.
type GetApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewGetApartmentUseCase створює новий use case.
func NewGetApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *GetApartmentUseCase {
	return &GetApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// GetApartmentInput - вхідні дані.
type GetApartmentInput struct {
	CurrentUserID int64
	ApartmentID   int64
}

// Execute виконує отримання квартири.
func (uc *GetApartmentUseCase) Execute(ctx context.Context, input GetApartmentInput) (*ApartmentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionRead,
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

	return mapApartmentToOutput(apartment), nil
}
