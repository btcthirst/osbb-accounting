package contractor

import (
	"context"
	"fmt"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

type DeleteContractorUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

func NewDeleteContractorUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteContractorUseCase {
	return &DeleteContractorUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteContractorInput struct {
	CurrentUserID int64
	ContractorID  int64
}

func (uc *DeleteContractorUseCase) Execute(ctx context.Context, input DeleteContractorInput) (*shared.DeleteOutput, error) {
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

	if err := uc.contractorRepo.SoftDelete(ctx, input.ContractorID); err != nil {
		return nil, fmt.Errorf("failed to delete contractor: %w", err)
	}

	return shared.NewDeleteOutput(
		fmt.Sprintf("Contractor '%s' successfully deleted", contractor.GetDisplayName()),
	), nil
}
