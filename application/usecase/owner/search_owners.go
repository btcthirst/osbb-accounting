// application/usecase/owner/search_owners.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// SearchOwnersUseCase шукає власників (для autocomplete).
type SearchOwnersUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewSearchOwnersUseCase створює новий use case.
func NewSearchOwnersUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *SearchOwnersUseCase {
	return &SearchOwnersUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// SearchOwnersInput - вхідні дані.
type SearchOwnersInput struct {
	CurrentUserID int64
	Query         string
	Limit         int
}

// Execute виконує швидкий пошук власників.
func (uc *SearchOwnersUseCase) Execute(ctx context.Context, input SearchOwnersInput) ([]*OwnerOutput, error) {
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

	// Мінімальна довжина запиту
	if len(input.Query) < 2 {
		return []*OwnerOutput{}, nil
	}

	// Ліміт за замовчуванням
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// Пошук
	owners, err := uc.ownerRepo.Search(ctx, input.Query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search owners: %w", err)
	}

	// Конвертація
	result := make([]*OwnerOutput, len(owners))
	for i, owner := range owners {
		result[i] = mapOwnerToOutput(owner)
	}

	return result, nil
}
