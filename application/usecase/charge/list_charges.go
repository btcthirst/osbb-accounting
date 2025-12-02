// application/usecase/charge/list_charges.go
package charge

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListChargesUseCase отримує список нарахувань.
type ListChargesUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewListChargesUseCase створює новий use case.
func NewListChargesUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *ListChargesUseCase {
	return &ListChargesUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// ListChargesInput - вхідні дані.
type ListChargesInput struct {
	CurrentUserID    int64
	IncludeDeleted   bool
	OwnershipShareID *int64
	ChargeType       *entity.ChargeType
	PeriodMonth      *int
	PeriodYear       *int
	MinAmount        *float64
	MaxAmount        *float64
	SearchQuery      string
	Limit            int
	Offset           int
	OrderBy          string
	OrderDesc        bool
}

// ListChargesOutput - результат.
type ListChargesOutput struct {
	Charges []*ChargeOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

// Execute виконує отримання списку нарахувань.
func (uc *ListChargesUseCase) Execute(ctx context.Context, input ListChargesInput) (*ListChargesOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.ChargeFilter{
		IncludeDeleted:   input.IncludeDeleted,
		OwnershipShareID: input.OwnershipShareID,
		ChargeType:       input.ChargeType,
		PeriodMonth:      input.PeriodMonth,
		PeriodYear:       input.PeriodYear,
		MinAmount:        input.MinAmount,
		MaxAmount:        input.MaxAmount,
		SearchQuery:      input.SearchQuery,
		Limit:            input.Limit,
		Offset:           input.Offset,
		OrderBy:          input.OrderBy,
		OrderDesc:        input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	charges, err := uc.chargeRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list charges: %w", err)
	}

	total, err := uc.chargeRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count charges: %w", err)
	}

	outputs := make([]*ChargeOutput, len(charges))
	for i, c := range charges {
		outputs[i] = mapChargeToOutput(c)
	}

	return &ListChargesOutput{
		Charges: outputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}
