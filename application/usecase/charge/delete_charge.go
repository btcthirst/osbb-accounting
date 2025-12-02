// application/usecase/charge/delete_charge.go
package charge

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// DeleteChargeUseCase видаляє нарахування.
type DeleteChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewDeleteChargeUseCase створює новий use case.
func NewDeleteChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteChargeUseCase {
	return &DeleteChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// DeleteChargeInput - вхідні дані.
type DeleteChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
}

// Execute виконує видалення нарахування.
func (uc *DeleteChargeUseCase) Execute(ctx context.Context, input DeleteChargeInput) (*shared.DeleteOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionDelete,
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

	if err := uc.chargeRepo.SoftDelete(ctx, input.ChargeID); err != nil {
		return nil, fmt.Errorf("failed to delete charge: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Charge for period %s successfully deleted", charge.GetPeriodDisplay()),
	), nil
}
