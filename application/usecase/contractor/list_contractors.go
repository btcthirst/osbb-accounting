// application/usecase/contractor/list_contractors.go
package contractor

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListContractorsUseCase отримує список контрагентів.
type ListContractorsUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

// NewListContractorsUseCase створює новий use case.
func NewListContractorsUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *ListContractorsUseCase {
	return &ListContractorsUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

// ListContractorsInput - вхідні дані.
type ListContractorsInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string
	ContractorType *entity.ContractorType
	HasBankDetails *bool
	HasContract    *bool
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

// ListContractorsOutput - результат.
type ListContractorsOutput struct {
	Contractors []*ContractorOutput
	Total       int64
	Limit       int
	Offset      int
	HasMore     bool
}

// Execute виконує отримання списку контрагентів.
func (uc *ListContractorsUseCase) Execute(ctx context.Context, input ListContractorsInput) (*ListContractorsOutput, error) {
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

	// Підготовка фільтру
	filter := repository.ContractorFilter{
		IncludeDeleted: input.IncludeDeleted,
		IsActive:       input.IsActive,
		SearchQuery:    input.SearchQuery,
		ContractorType: input.ContractorType,
		HasBankDetails: input.HasBankDetails,
		HasContract:    input.HasContract,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку
	contractors, err := uc.contractorRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list contractors: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.contractorRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count contractors: %w", err)
	}

	// Конвертація в output
	contractorOutputs := make([]*ContractorOutput, len(contractors))
	for i, c := range contractors {
		contractorOutputs[i] = mapContractorToOutput(c)
	}

	return &ListContractorsOutput{
		Contractors: contractorOutputs,
		Total:       total,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
		HasMore:     int64(filter.Offset+filter.Limit) < total,
	}, nil
}
