// application/usecase/charge/get_charge.go
package charge

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetChargeUseCase отримує нарахування за ID.
type GetChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewGetChargeUseCase створює новий use case.
func NewGetChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *GetChargeUseCase {
	return &GetChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// GetChargeInput - вхідні дані.
type GetChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
}

// Execute виконує отримання нарахування.
func (uc *GetChargeUseCase) Execute(ctx context.Context, input GetChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	charge, err := uc.chargeRepo.GetByID(ctx, input.ChargeID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"charge not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}
