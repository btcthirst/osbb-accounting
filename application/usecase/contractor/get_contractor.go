package contractor

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

type GetContractorUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

func NewGetContractorUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *GetContractorUseCase {
	return &GetContractorUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

type GetContractorInput struct {
	CurrentUserID int64
	ContractorID  int64
}

func (uc *GetContractorUseCase) Execute(ctx context.Context, input GetContractorInput) (*ContractorOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "contractors", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	contractor, err := uc.contractorRepo.GetByID(ctx, input.ContractorID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"contractor not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get contractor: %w", err)
	}

	return mapContractorToOutput(contractor), nil
}
