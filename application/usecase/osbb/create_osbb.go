// application/usecase/osbb/create_osbb.go
package osbb

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateOSBBUseCase створює організацію ОСББ.
type CreateOSBBUseCase struct {
	osbbRepo       repository.OSBBRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateOSBBUseCase створює новий use case.
func NewCreateOSBBUseCase(
	osbbRepo repository.OSBBRepository,
	permissionRepo repository.PermissionRepository,
) *CreateOSBBUseCase {
	return &CreateOSBBUseCase{
		osbbRepo:       osbbRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateOSBBInput - вхідні дані.
type CreateOSBBInput struct {
	CurrentUserID int64
	Name          string
	EDRPOU        string
	LegalAddress  string
	ActualAddress *string
	Phone         *string
	Email         *string
	Website       *string
	ChairmanName  string
}

// Execute виконує створення ОСББ.
func (uc *CreateOSBBUseCase) Execute(ctx context.Context, input CreateOSBBInput) (*GetOSBBOutput, error) {
	// Перевірка дозволів (тільки admin)
	hasPermission, err := uc.permissionRepo.HasPermission(ctx, input.CurrentUserID, "system.all")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"only administrators can create OSBB organization",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Перевірка чи вже існує ОСББ
	exists, err := uc.osbbRepo.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check OSBB existence: %w", err)
	}
	if exists {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeAlreadyExists,
			"OSBB organization already exists",
			domainErrors.ErrAlreadyExists,
		)
	}

	// Створення entity
	osbb, err := entity.NewOSBB(
		input.Name,
		input.EDRPOU,
		input.LegalAddress,
		input.ChairmanName,
		input.ActualAddress,
		input.Phone,
		input.Email,
		input.Website,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid OSBB data",
			err,
		)
	}

	// Збереження
	if err := uc.osbbRepo.Create(ctx, osbb); err != nil {
		return nil, fmt.Errorf("failed to create OSBB: %w", err)
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
