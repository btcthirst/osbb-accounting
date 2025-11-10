// application/usecase/auth/login_user.go
package auth

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
	"osbb-accounting/domain/service"
	"osbb-accounting/domain/valueobject"
)

// LoginUserUseCase обробляє вхід користувача в систему.
type LoginUserUseCase struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	passwordHasher service.PasswordHasher
}

// NewLoginUserUseCase створює новий use case.
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	passwordHasher service.PasswordHasher,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		passwordHasher: passwordHasher,
	}
}

// LoginUserInput - вхідні дані для входу.
type LoginUserInput struct {
	UsernameOrEmail string
	Password        string
	IPAddress       *string
	UserAgent       *string
	SessionDuration time.Duration // Опціонально - час життя сесії
}

// LoginUserOutput - результат входу.
type LoginUserOutput struct {
	SessionToken string
	ExpiresAt    time.Time
	User         UserInfo
	Roles        []string
	Permissions  []string
}

// UserInfo - інформація про користувача.
type UserInfo struct {
	ID        int64
	Username  string
	Email     string
	FirstName string
	LastName  string
	FullName  string
}

// Execute виконує вхід користувача.
func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
	// 1. Валідація вхідних даних через Value Object
	credentials, err := valueobject.NewCredentials(input.UsernameOrEmail, input.Password)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidCredentials,
			"invalid credentials format",
			err,
		)
	}

	// 2. Пошук користувача
	user, err := uc.userRepo.GetByUsernameOrEmail(ctx, credentials.UsernameOrEmail)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeInvalidCredentials,
				"invalid username/email or password",
				domainErrors.ErrInvalidCredentials,
			)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 3. Перевірка чи може користувач увійти
	if err := user.CanLogin(); err != nil {
		if domainErrors.Is(err, entity.ErrUserDeleted) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeAuthenticationFailed,
				"user account is deleted",
				err,
			)
		}
		if domainErrors.Is(err, entity.ErrUserInactive) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeAuthenticationFailed,
				"user account is inactive",
				err,
			)
		}
		return nil, fmt.Errorf("user cannot login: %w", err)
	}

	// 4. Перевірка пароля
	if err := uc.passwordHasher.Verify(credentials.Password, user.PasswordHash); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidCredentials,
			"invalid username/email or password",
			domainErrors.ErrInvalidCredentials,
		)
	}

	// 5. Створення сесії
	sessionDuration := input.SessionDuration
	if sessionDuration == 0 {
		sessionDuration = entity.SessionDuration // 24 години за замовчуванням
	}

	session, err := entity.NewSession(user.ID, sessionDuration, input.IPAddress, input.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// 6. Оновлення часу останнього входу
	user.RecordLogin()
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Не критична помилка - сесія вже створена
		// Можна залогувати для моніторингу
	}

	// 7. Отримання ролей користувача
	roles, err := uc.roleRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// 8. Отримання дозволів користувача
	permissions, err := uc.permissionRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permissionCodes := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionCodes[i] = perm.Code
	}

	// 9. Формування відповіді
	return &LoginUserOutput{
		SessionToken: session.Token,
		ExpiresAt:    session.ExpiresAt,
		User: UserInfo{
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

// Validate перевіряє вхідні дані перед виконанням.
func (input *LoginUserInput) Validate() error {
	validationErrors := domainErrors.NewValidationErrors()

	if input.UsernameOrEmail == "" {
		validationErrors.Add("username_or_email", "username or email is required", input.UsernameOrEmail)
	}
	if input.Password == "" {
		validationErrors.Add("password", "password is required", input.Password)
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
