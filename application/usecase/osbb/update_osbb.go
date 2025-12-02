// application/usecase/osbb/update_osbb.go
package osbb

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateOSBBUseCase оновлює дані ОСББ.
type UpdateOSBBUseCase struct {
	osbbRepo       repository.OSBBRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateOSBBUseCase створює новий use case.
func NewUpdateOSBBUseCase(
	osbbRepo repository.OSBBRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateOSBBUseCase {
	return &UpdateOSBBUseCase{
		osbbRepo:       osbbRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateOSBBInput - вхідні дані.
type UpdateOSBBInput struct {
	CurrentUserID int64
	Name          string
	LegalAddress  string
	ActualAddress *string
	Phone         *string
	Email         *string
	Website       *string
	ChairmanName  string
}

// Execute виконує оновлення ОСББ.
func (uc *UpdateOSBBUseCase) Execute(ctx context.Context, input UpdateOSBBInput) (*GetOSBBOutput, error) {
	// Перевірка дозволів (admin або accountant)
	hasPermission, err := uc.permissionRepo.HasPermission(ctx, input.CurrentUserID, "system.all")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"insufficient permissions to update OSBB",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Отримання поточних даних
	osbb, err := uc.osbbRepo.Get(ctx)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"OSBB organization not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get OSBB: %w", err)
	}

	// Оновлення даних
	if err := osbb.Update(
		input.Name,
		input.LegalAddress,
		input.ChairmanName,
		input.ActualAddress,
		input.Phone,
		input.Email,
		input.Website,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid OSBB data",
			err,
		)
	}

	// Збереження
	if err := uc.osbbRepo.Update(ctx, osbb); err != nil {
		return nil, fmt.Errorf("failed to update OSBB: %w", err)
	}

	return &GetOSBBOutput{
		ID:            osbb.ID,
		Name:          osbb.Name,
		EDRPOU:        osbb.EDRPOU,
		LegalAddress:  osbb.LegalAddress,
		ActualAddress: osbb.ActualAddress,
		Phone:         osbb.Phone,
		Email:         osbb.Email,
		Website:       osbb.Website,
		ChairmanName:  osbb.ChairmanName,
	}, nil
}
