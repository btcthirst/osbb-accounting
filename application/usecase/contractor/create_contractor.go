// application/usecase/contractor/create_contractor.go
package contractor

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateContractorUseCase створює нового контрагента.
type CreateContractorUseCase struct {
	contractorRepo repository.ContractorRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateContractorUseCase створює новий use case.
func NewCreateContractorUseCase(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *CreateContractorUseCase {
	return &CreateContractorUseCase{
		contractorRepo: contractorRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateContractorInput - вхідні дані.
type CreateContractorInput struct {
	CurrentUserID  int64
	Name           string
	EDRPOU         *string
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

// Execute виконує створення контрагента.
func (uc *CreateContractorUseCase) Execute(ctx context.Context, input CreateContractorInput) (*ContractorOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "contractors", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"insufficient permissions to create contractor",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Перевірка унікальності ЄДРПОУ (якщо вказано)
	if input.EDRPOU != nil && *input.EDRPOU != "" {
		exists, err := uc.contractorRepo.ExistsByEDRPOU(ctx, *input.EDRPOU)
		if err != nil {
			return nil, fmt.Errorf("failed to check EDRPOU: %w", err)
		}
		if exists {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"contractor with this EDRPOU already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "edrpou")
		}
	}

	// Створення entity
	contractor, err := entity.NewContractor(
		input.Name,
		input.ContractorType,
		input.EDRPOU,
		input.ContactPerson,
		input.Phone,
		input.Email,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid contractor data",
			err,
		)
	}

	// Додаткові поля
	contractor.Address = input.Address
	contractor.BankAccount = input.BankAccount
	contractor.BankName = input.BankName
	contractor.BankMFO = input.BankMFO
	contractor.ContractNumber = input.ContractNumber
	contractor.ContractDate = input.ContractDate
	contractor.Notes = input.Notes

	// Збереження
	if err := uc.contractorRepo.Create(ctx, contractor); err != nil {
		return nil, fmt.Errorf("failed to create contractor: %w", err)
	}

	return mapContractorToOutput(contractor), nil
}
