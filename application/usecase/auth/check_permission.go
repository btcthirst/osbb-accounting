// application/usecase/auth/check_permission.go
package auth

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CheckPermissionUseCase перевіряє чи має користувач дозвіл.
type CheckPermissionUseCase struct {
	permissionRepo repository.PermissionRepository
}

// NewCheckPermissionUseCase створює новий use case.
func NewCheckPermissionUseCase(permissionRepo repository.PermissionRepository) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{
		permissionRepo: permissionRepo,
	}
}

// CheckPermissionInput - вхідні дані для перевірки дозволу.
type CheckPermissionInput struct {
	UserID         int64
	PermissionCode string // Напр. "owners.create"
}

// CheckPermissionOutput - результат перевірки.
type CheckPermissionOutput struct {
	Allowed bool
	Reason  string // Чому дозволено/заборонено (для дебагу)
}

// Execute перевіряє чи має користувач конкретний дозвіл.
func (uc *CheckPermissionUseCase) Execute(ctx context.Context, input CheckPermissionInput) (*CheckPermissionOutput, error) {
	// Валідація
	if input.UserID <= 0 {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"invalid user ID",
			domainErrors.ErrInvalidInput,
		)
	}
	if input.PermissionCode == "" {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"permission code is required",
			domainErrors.ErrInvalidInput,
		)
	}

	// Перевірка дозволу
	hasPermission, err := uc.permissionRepo.HasPermission(ctx, input.UserID, input.PermissionCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	if hasPermission {
		return &CheckPermissionOutput{
			Allowed: true,
			Reason:  fmt.Sprintf("user has permission '%s'", input.PermissionCode),
		}, nil
	}

	return &CheckPermissionOutput{
		Allowed: false,
		Reason:  fmt.Sprintf("user does not have permission '%s'", input.PermissionCode),
	}, nil
}

// CheckResourcePermissionInput - вхідні дані для перевірки дозволу на ресурс.
type CheckResourcePermissionInput struct {
	UserID   int64
	Resource string        // Напр. "owners"
	Action   entity.Action // Напр. ActionCreate
}

// ExecuteResourceCheck перевіряє чи має користувач дозвіл на дію над ресурсом.
func (uc *CheckPermissionUseCase) ExecuteResourceCheck(ctx context.Context, input CheckResourcePermissionInput) (*CheckPermissionOutput, error) {
	// Валідація
	if input.UserID <= 0 {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"invalid user ID",
			domainErrors.ErrInvalidInput,
		)
	}
	if input.Resource == "" {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"resource is required",
			domainErrors.ErrInvalidInput,
		)
	}
	if !input.Action.IsValid() {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"invalid action",
			domainErrors.ErrInvalidInput,
		)
	}

	// Перевірка дозволу
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx,
		input.UserID,
		input.Resource,
		input.Action,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check resource permission: %w", err)
	}

	if hasPermission {
		return &CheckPermissionOutput{
			Allowed: true,
			Reason:  fmt.Sprintf("user has permission for '%s.%s'", input.Resource, input.Action),
		}, nil
	}

	return &CheckPermissionOutput{
		Allowed: false,
		Reason:  fmt.Sprintf("user does not have permission for '%s.%s'", input.Resource, input.Action),
	}, nil
}

// RequirePermission - helper функція для use cases, яка викидає помилку якщо немає дозволу.
func (uc *CheckPermissionUseCase) RequirePermission(ctx context.Context, userID int64, permissionCode string) error {
	output, err := uc.Execute(ctx, CheckPermissionInput{
		UserID:         userID,
		PermissionCode: permissionCode,
	})
	if err != nil {
		return err
	}

	if !output.Allowed {
		return domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			output.Reason,
			domainErrors.ErrPermissionDenied,
		).WithDetails("permission", permissionCode)
	}

	return nil
}

// RequireResourcePermission - helper для перевірки дозволу на ресурс.
func (uc *CheckPermissionUseCase) RequireResourcePermission(
	ctx context.Context,
	userID int64,
	resource string,
	action entity.Action,
) error {
	output, err := uc.ExecuteResourceCheck(ctx, CheckResourcePermissionInput{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	})
	if err != nil {
		return err
	}

	if !output.Allowed {
		return domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			output.Reason,
			domainErrors.ErrPermissionDenied,
		).WithDetails("resource", resource).WithDetails("action", string(action))
	}

	return nil
}
