package ownership

import (
	"context"
	"fmt"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// DeleteOwnershipShareUseCase
// ============================================================================

func NewDeleteOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteOwnershipShareUseCase {
	return &DeleteOwnershipShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteOwnershipShareInput struct {
	CurrentUserID int64
	ShareID       int64
}

func (uc *DeleteOwnershipShareUseCase) Execute(
	ctx context.Context,
	input DeleteOwnershipShareInput,
) (*shared.DeleteOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання для інформації
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

	// Видалення
	if err := uc.shareRepo.SoftDelete(ctx, input.ShareID); err != nil {
		return nil, fmt.Errorf("failed to delete ownership share: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Ownership share deleted: %s - %s (%s)",
			details.OwnerName,
			details.ApartmentNumber,
			details.Share.GetShareDisplay()),
	), nil
}
