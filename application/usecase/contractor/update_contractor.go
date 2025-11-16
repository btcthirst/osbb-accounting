// application/usecase/contractor/update_contractor.go
package contractor

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateContractorUseCase оновлює дані контрагента.
type UpdateContractorUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateContractorUseCase створює новий use case.
func NewUpdateContractorUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateContractorUseCase {
	return &UpdateContractorUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateContractorInput - вхідні дані.
type UpdateContractorInput struct {
	CurrentUserID  int64
	ContractorID   int64
	Name           string
	ContractorType entity.ContractorType
	ContactPerson  *string
	Phone          *string
	Email          *string
	Address        *string
	BankAccount    *string
	BankName       *string
	BankMFO        *string
	ContractNumber *string
	ContractDate   *time.Time
	Notes          *string
}

// Execute виконує оновлення контрагента.
func (uc *UpdateContractorUseCase) Execute(ctx context.Context, input UpdateContractorInput) (*ContractorOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "contractors", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання поточних даних
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

	// Оновлення основних полів
	if err := contractor.Update(
		input.Name,
		input.ContractorType,
		input.ContactPerson,
		input.Phone,
		input.Email,
		input.Address,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid contractor data",
			err,
		)
	}

	// Оновлення банківських реквізитів
	if err := contractor.UpdateBankDetails(
		input.BankAccount,
		input.BankName,
		input.BankMFO,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid bank details",
			err,
		)
	}

	// Оновлення договірної інформації
	if err := contractor.UpdateContract(
		input.ContractNumber,
		input.ContractDate,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid contract details",
			err,
		)
	}

	// Збереження
	if err := uc.contractorRepo.Update(ctx, contractor); err != nil {
		return nil, fmt.Errorf("failed to update contractor: %w", err)
	}

	return mapContractorToOutput(contractor), nil
}
