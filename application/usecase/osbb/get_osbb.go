// application/usecase/osbb/get_osbb.go
package osbb

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
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

// GetOSBBOutput - результат.
type GetOSBBOutput struct {
	ID            int64
	Name          string
	EDRPOU        string
	LegalAddress  string
	ActualAddress *string
	Phone         *string
	Email         *string
	Website       *string
	ChairmanName  string
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

// ============================================================================

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

// ============================================================================

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
