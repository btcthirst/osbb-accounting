// application/usecase/charge/update_charge.go
package charge

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateChargeUseCase оновлює нарахування.
type UpdateChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateChargeUseCase створює новий use case.
func NewUpdateChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateChargeUseCase {
	return &UpdateChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateChargeInput - вхідні дані.
type UpdateChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
	Amount        float64
	Tariff        *float64
	Quantity      *float64
	Description   *string
	Notes         *string
}

// Execute виконує оновлення нарахування.
func (uc *UpdateChargeUseCase) Execute(ctx context.Context, input UpdateChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionUpdate,
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

	if err := charge.Update(
		input.Amount,
		input.Tariff,
		input.Quantity,
		input.Description,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid charge data",
			err,
		)
	}

	if err := uc.chargeRepo.Update(ctx, charge); err != nil {
		return nil, fmt.Errorf("failed to update charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}
