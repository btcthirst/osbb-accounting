package contractor

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

type SearchContractorsUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

func NewSearchContractorsUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *SearchContractorsUseCase {
	return &SearchContractorsUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

type SearchContractorsInput struct {
	CurrentUserID int64
	Query         string
	Limit         int
}

func (uc *SearchContractorsUseCase) Execute(ctx context.Context, input SearchContractorsInput) ([]*ContractorOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "contractors", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Мінімальна довжина запиту
	if len(input.Query) < 2 {
		return []*ContractorOutput{}, nil
	}

	// Ліміт за замовчуванням
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// Пошук
	contractors, err := uc.contractorRepo.Search(ctx, input.Query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search contractors: %w", err)
	}

	// Конвертація
	result := make([]*ContractorOutput, len(contractors))
	for i, contractor := range contractors {
		result[i] = mapContractorToOutput(contractor)
	}

	return result, nil
}
