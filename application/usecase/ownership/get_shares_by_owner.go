package ownership

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetByOwnerUseCase - отримати всі квартири власника
// ============================================================================

type GetByOwnerUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewGetByOwnerUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *GetByOwnerUseCase {
	return &GetByOwnerUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type GetByOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
	OnlyActive    bool
}

func (uc *GetByOwnerUseCase) Execute(
	ctx context.Context,
	input GetByOwnerInput,
) ([]*OwnershipShareDetailsOutput, error) {
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

	// Фільтр
	filter := repository.OwnershipShareFilter{
		OwnerID:           &input.OwnerID,
		IsCurrentlyActive: input.OnlyActive,
		Limit:             1000,
		OrderBy:           "start_date",
		OrderDesc:         true,
	}

	// Отримання
	detailsList, err := uc.shareRepo.ListWithDetails(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ownership shares: %w", err)
	}

	// Конвертація
	outputs := make([]*OwnershipShareDetailsOutput, len(detailsList))
	for i, details := range detailsList {
		outputs[i] = mapDetailsToOutput(details)
	}

	return outputs, nil
}
