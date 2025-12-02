// application/usecase/charge/create_charge.go
package charge

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateChargeUseCase створює нове нарахування.
type CreateChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateChargeUseCase створює новий use case.
func NewCreateChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *CreateChargeUseCase {
	return &CreateChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateChargeInput - вхідні дані.
type CreateChargeInput struct {
	CurrentUserID    int64
	OwnershipShareID int64
	ChargeType       entity.ChargeType
	ChargeDate       time.Time
	PeriodMonth      int
	PeriodYear       int
	Amount           float64
	Tariff           *float64
	Quantity         *float64
	Description      *string
	Notes            *string
}

// Execute виконує створення нарахування.
func (uc *CreateChargeUseCase) Execute(ctx context.Context, input CreateChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	charge, err := entity.NewCharge(
		input.OwnershipShareID,
		input.ChargeType,
		input.ChargeDate,
		input.PeriodMonth,
		input.PeriodYear,
		input.Amount,
		input.Tariff,
		input.Quantity,
		input.Description,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid charge data",
			err,
		)
	}

	charge.Notes = input.Notes

	if err := uc.chargeRepo.Create(ctx, charge); err != nil {
		return nil, fmt.Errorf("failed to create charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}
