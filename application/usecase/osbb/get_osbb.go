// application/usecase/osbb/get_osbb.go
package osbb

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetOSBBUseCase отримує дані організації ОСББ.
type GetOSBBUseCase struct {
	osbbRepo       repository.OSBBRepository
	permissionRepo repository.PermissionRepository
}

// NewGetOSBBUseCase створює новий use case.
func NewGetOSBBUseCase(
	osbbRepo repository.OSBBRepository,
	permissionRepo repository.PermissionRepository,
) *GetOSBBUseCase {
	return &GetOSBBUseCase{
		osbbRepo:       osbbRepo,
		permissionRepo: permissionRepo,
	}
}

// GetOSBBInput - вхідні дані.
type GetOSBBInput struct {
	CurrentUserID int64
}

// Execute виконує отримання даних ОСББ.
func (uc *GetOSBBUseCase) Execute(ctx context.Context, input GetOSBBInput) (*GetOSBBOutput, error) {
	// Перевірка дозволів (всі можуть переглядати ОСББ)
	// Для більш суворої моделі можна додати перевірку:
	// hasPermission, _ := uc.permissionRepo.HasPermissionForResource(ctx, input.CurrentUserID, "osbb", entity.ActionRead)
	// if !hasPermission { return nil, domainErrors.ErrPermissionDenied }

	// Отримання ОСББ
	osbb, err := uc.osbbRepo.Get(ctx)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"OSBB organization not found. Please initialize it first.",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get OSBB: %w", err)
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
