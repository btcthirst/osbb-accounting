package ownership

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetOwnershipShareUseCase
// ============================================================================

type GetOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewGetOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *GetOwnershipShareUseCase {
	return &GetOwnershipShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type GetOwnershipShareInput struct {
	CurrentUserID int64
	ShareID       int64
}

func (uc *GetOwnershipShareUseCase) Execute(
	ctx context.Context,
	input GetOwnershipShareInput,
) (*OwnershipShareDetailsOutput, error) {
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

	// Отримання з деталями
	details, err := uc.shareRepo.GetWithDetails(ctx, input.ShareID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"ownership share not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get ownership share: %w", err)
	}

	return mapDetailsToOutput(details), nil
}
