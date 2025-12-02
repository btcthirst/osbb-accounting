// application/usecase/owner/create_owner_new.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateOwnerUseCase створює нового власника.
type CreateOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateOwnerUseCase створює новий use case.
func NewCreateOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *CreateOwnerUseCase {
	return &CreateOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateOwnerInput - вхідні дані.
type CreateOwnerInput struct {
	CurrentUserID     int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
}


// Execute виконує створення власника.
func (uc *CreateOwnerUseCase) Execute(ctx context.Context, input CreateOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"insufficient permissions to create owner",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Перевірка унікальності ІПН (якщо вказано)
	if input.TaxNumber != nil && *input.TaxNumber != "" {
		exists, err := uc.ownerRepo.ExistsByTaxNumber(ctx, *input.TaxNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to check tax number: %w", err)
		}
		if exists {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"owner with this tax number already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "tax_number")
		}
	}

	// Створення entity
	owner, err := entity.NewOwner(
		input.FirstName,
		input.LastName,
		input.MiddleName,
		input.Phone,
		input.Email,
		input.TaxNumber,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid owner data",
			err,
		)
	}

	// Додаткові поля
	owner.PassportSeries = input.PassportSeries
	owner.PassportNumber = input.PassportNumber
	owner.RegisteredAddress = input.RegisteredAddress
	owner.ActualAddress = input.ActualAddress
	owner.Notes = input.Notes

	// Збереження
	if err := uc.ownerRepo.Create(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to create owner: %w", err)
	}

	return mapOwnerToOutput(owner), nil
}
