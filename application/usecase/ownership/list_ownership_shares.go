package ownership

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// ListOwnershipSharesUseCase
// ============================================================================

type ListOwnershipSharesUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewListOwnershipSharesUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *ListOwnershipSharesUseCase {
	return &ListOwnershipSharesUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type ListOwnershipSharesInput struct {
	CurrentUserID     int64
	IncludeDeleted    bool
	IsActive          *bool
	OwnerID           *int64
	ApartmentID       *int64
	OwnershipType     *entity.OwnershipType
	IsCurrentlyActive bool
	Limit             int
	Offset            int
	OrderBy           string
	OrderDesc         bool
}

type ListOwnershipSharesOutput struct {
	Shares  []*OwnershipShareDetailsOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

func (uc *ListOwnershipSharesUseCase) Execute(
	ctx context.Context,
	input ListOwnershipSharesInput,
) (*ListOwnershipSharesOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Підготовка фільтру
	filter := repository.OwnershipShareFilter{
		IncludeDeleted:    input.IncludeDeleted,
		IsActive:          input.IsActive,
		OwnerID:           input.OwnerID,
		ApartmentID:       input.ApartmentID,
		OwnershipType:     input.OwnershipType,
		IsCurrentlyActive: input.IsCurrentlyActive,
		Limit:             input.Limit,
		Offset:            input.Offset,
		OrderBy:           input.OrderBy,
		OrderDesc:         input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку з деталями
	detailsList, err := uc.shareRepo.ListWithDetails(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list ownership shares: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.shareRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count ownership shares: %w", err)
	}

	// Конвертація в output
	outputs := make([]*OwnershipShareDetailsOutput, len(detailsList))
	for i, details := range detailsList {
		outputs[i] = mapDetailsToOutput(details)
	}

	return &ListOwnershipSharesOutput{
		Shares:  outputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}
