// application/usecase/owner/get_owner.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetOwnerUseCase отримує власника за ID.
type GetOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewGetOwnerUseCase створює новий use case.
func NewGetOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *GetOwnerUseCase {
	return &GetOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// GetOwnerInput - вхідні дані.
type GetOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
}

// Execute виконує отримання власника.
func (uc *GetOwnerUseCase) Execute(ctx context.Context, input GetOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання власника
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

	return mapOwnerToOutput(owner), nil
}
