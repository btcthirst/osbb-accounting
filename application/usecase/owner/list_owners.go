// application/usecase/owner/list_owners.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListOwnersUseCase отримує список власників.
type ListOwnersUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewListOwnersUseCase створює новий use case.
func NewListOwnersUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *ListOwnersUseCase {
	return &ListOwnersUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// ListOwnersInput - вхідні дані.
type ListOwnersInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string
	HasTaxNumber   *bool
	HasContact     *bool
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

// ListOwnersOutput - результат.
type ListOwnersOutput struct {
	Owners  []*OwnerOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

// Execute виконує отримання списку власників.
func (uc *ListOwnersUseCase) Execute(ctx context.Context, input ListOwnersInput) (*ListOwnersOutput, error) {
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

	// Підготовка фільтру
	filter := repository.OwnerFilter{
		IncludeDeleted: input.IncludeDeleted,
		IsActive:       input.IsActive,
		SearchQuery:    input.SearchQuery,
		HasTaxNumber:   input.HasTaxNumber,
		HasContact:     input.HasContact,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	// За замовчуванням
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку
	owners, err := uc.ownerRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.ownerRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count owners: %w", err)
	}

	// Конвертація в output
	ownerOutputs := make([]*OwnerOutput, len(owners))
	for i, owner := range owners {
		ownerOutputs[i] = mapOwnerToOutput(owner)
	}

	return &ListOwnersOutput{
		Owners:  ownerOutputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}
