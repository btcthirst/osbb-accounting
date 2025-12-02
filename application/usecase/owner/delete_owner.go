// application/usecase/owner/delete_owner.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// DeleteOwnerUseCase видаляє (деактивує) власника.
type DeleteOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewDeleteOwnerUseCase створює новий use case.
func NewDeleteOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteOwnerUseCase {
	return &DeleteOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// DeleteOwnerInput - вхідні дані.
type DeleteOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
}

// Execute виконує видалення власника.
func (uc *DeleteOwnerUseCase) Execute(ctx context.Context, input DeleteOwnerInput) (*shared.DeleteOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Перевірка існування
	owner, err := uc.ownerRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"owner not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// Soft delete
	if err := uc.ownerRepo.SoftDelete(ctx, input.OwnerID); err != nil {
		return nil, fmt.Errorf("failed to delete owner: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Owner '%s' successfully deleted", owner.FullName()),
	), nil
}
