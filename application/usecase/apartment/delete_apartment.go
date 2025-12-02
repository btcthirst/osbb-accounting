// application/usecase/apartment/delete_apartment.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// DeleteApartmentUseCase видаляє квартиру.
type DeleteApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewDeleteApartmentUseCase створює новий use case.
func NewDeleteApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteApartmentUseCase {
	return &DeleteApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// DeleteApartmentInput - вхідні дані.
type DeleteApartmentInput struct {
	CurrentUserID int64
	ApartmentID   int64
}

// Execute виконує видалення квартири.
func (uc *DeleteApartmentUseCase) Execute(ctx context.Context, input DeleteApartmentInput) (*shared.DeleteOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionDelete,
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

	if err := uc.apartmentRepo.SoftDelete(ctx, input.ApartmentID); err != nil {
		return nil, fmt.Errorf("failed to delete apartment: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Apartment '%s' successfully deleted", apartment.GetDisplayName()),
	), nil
}
