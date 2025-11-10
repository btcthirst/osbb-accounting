// application/usecase/auth/validate_session.go
package auth

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ValidateSessionUseCase перевіряє валідність сесії.
type ValidateSessionUseCase struct {
	sessionRepo    repository.SessionRepository
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewValidateSessionUseCase створює новий use case.
func NewValidateSessionUseCase(
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *ValidateSessionUseCase {
	return &ValidateSessionUseCase{
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// ValidateSessionInput - вхідні дані для валідації.
type ValidateSessionInput struct {
	SessionToken string
}

// ValidateSessionOutput - результат валідації.
type ValidateSessionOutput struct {
	Valid       bool
	User        *UserInfo
	Roles       []string
	Permissions []string
}

// Execute перевіряє валідність сесії та повертає інформацію про користувача.
func (uc *ValidateSessionUseCase) Execute(ctx context.Context, input ValidateSessionInput) (*ValidateSessionOutput, error) {
	// 1. Валідація токену
	if input.SessionToken == "" {
		return &ValidateSessionOutput{Valid: false}, domainErrors.NewDomainError(
			domainErrors.CodeInvalidToken,
			"session token is required",
			domainErrors.ErrInvalidToken,
		)
	}

	// 2. Пошук сесії
	session, err := uc.sessionRepo.GetByToken(ctx, input.SessionToken)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrSessionNotFound) {
			return &ValidateSessionOutput{Valid: false}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 3. Перевірка чи не закінчилася сесія
	if err := session.IsValid(); err != nil {
		// Сесія закінчилася - можна видалити її
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return &ValidateSessionOutput{Valid: false}, nil
	}

	// 4. Отримання користувача
	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			// Користувач не існує - видаляємо сесію
			_ = uc.sessionRepo.Delete(ctx, session.ID)
			return &ValidateSessionOutput{Valid: false}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 5. Перевірка чи може користувач використовувати систему
	if err := user.CanLogin(); err != nil {
		// Користувач деактивований або видалений - видаляємо сесію
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return &ValidateSessionOutput{Valid: false}, nil
	}

	// 6. Отримання ролей
	roles, err := uc.roleRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// 7. Отримання дозволів
	permissions, err := uc.permissionRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permissionCodes := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionCodes[i] = perm.Code
	}

	// 8. Формування відповіді
	return &ValidateSessionOutput{
		Valid: true,
		User: &UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			FullName:  user.FullName(),
		},
		Roles:       roleNames,
		Permissions: permissionCodes,
	}, nil
}
