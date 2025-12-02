// application/usecase/owner/update_owner.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateOwnerUseCase оновлює дані власника.
type UpdateOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateOwnerUseCase створює новий use case.
func NewUpdateOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateOwnerUseCase {
	return &UpdateOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateOwnerInput - вхідні дані.
type UpdateOwnerInput struct {
	CurrentUserID     int64
	OwnerID           int64
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

// Execute виконує оновлення власника.
func (uc *UpdateOwnerUseCase) Execute(ctx context.Context, input UpdateOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання поточних даних
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

	// Перевірка унікальності ІПН (якщо змінився)
	if input.TaxNumber != nil && *input.TaxNumber != "" {
		if owner.TaxNumber == nil || *owner.TaxNumber != *input.TaxNumber {
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
	}

	// Оновлення основних полів
	if err := owner.Update(
		input.FirstName,
		input.LastName,
		input.MiddleName,
		input.Phone,
		input.Email,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid owner data",
			err,
		)
	}

	// Оновлення ІПН
	if err := owner.UpdateTaxNumber(input.TaxNumber); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid tax number",
			err,
		)
	}

	// Оновлення паспорту та адрес
	owner.UpdatePassport(input.PassportSeries, input.PassportNumber)
	owner.UpdateAddresses(input.RegisteredAddress, input.ActualAddress)
	owner.Notes = input.Notes

	// Збереження
	if err := uc.ownerRepo.Update(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to update owner: %w", err)
	}

	return mapOwnerToOutput(owner), nil
}
