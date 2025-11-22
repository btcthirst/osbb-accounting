// application/service/auth_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"osbb-accounting/domain/service"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, input auth.RegisterUserInput) (*auth.RegisterUserOutput, error)
	Login(ctx context.Context, input auth.LoginUserInput) (*auth.LoginUserOutput, error)
	Logout(ctx context.Context, input auth.LogoutUserInput) (*auth.LogoutUserOutput, error)
	LogoutAll(ctx context.Context, input auth.LogoutAllUserSessionsInput) (*auth.LogoutUserOutput, error)
	ValidateSession(ctx context.Context, input auth.ValidateSessionInput) (*auth.ValidateSessionOutput, error)
	ChangePassword(ctx context.Context, input auth.ChangePasswordInput) (*auth.ChangePasswordOutput, error)
	CheckPermission(ctx context.Context, input auth.CheckPermissionInput) (*auth.CheckPermissionOutput, error)
	CheckResourcePermission(ctx context.Context, input auth.CheckResourcePermissionInput) (*auth.CheckPermissionOutput, error)
	RequirePermission(ctx context.Context, userID int64, permissionCode string) error
	RequireResourcePermission(ctx context.Context, userID int64, resource string, action entity.Action) error
}

// AuthService - фасад для всіх операцій аутентифікації.
// Об'єднує всі use cases в один інтерфейс для зручності використання.
type AuthService struct {
	registerUser    *auth.RegisterUserUseCase
	loginUser       *auth.LoginUserUseCase
	logoutUser      *auth.LogoutUserUseCase
	validateSession *auth.ValidateSessionUseCase
	changePassword  *auth.ChangePasswordUseCase
	checkPermission *auth.CheckPermissionUseCase
}

// NewAuthService створює новий AuthService.
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	passwordHasher service.PasswordHasher,
) *AuthService {
	return &AuthService{
		registerUser: auth.NewRegisterUserUseCase(
			userRepo,
			roleRepo,
			passwordHasher,
		),
		loginUser: auth.NewLoginUserUseCase(
			userRepo,
			sessionRepo,
			roleRepo,
			permissionRepo,
			passwordHasher,
		),
		logoutUser: auth.NewLogoutUserUseCase(
			sessionRepo,
		),
		validateSession: auth.NewValidateSessionUseCase(
			sessionRepo,
			userRepo,
			roleRepo,
			permissionRepo,
		),
		changePassword: auth.NewChangePasswordUseCase(
			userRepo,
			sessionRepo,
			passwordHasher,
		),
		checkPermission: auth.NewCheckPermissionUseCase(
			permissionRepo,
		),
	}
}

// Register реєструє нового користувача.
func (s *AuthService) Register(ctx context.Context, input auth.RegisterUserInput) (*auth.RegisterUserOutput, error) {
	return s.registerUser.Execute(ctx, input)
}

// Login виконує вхід користувача.
func (s *AuthService) Login(ctx context.Context, input auth.LoginUserInput) (*auth.LoginUserOutput, error) {
	return s.loginUser.Execute(ctx, input)
}

// Logout виконує вихід користувача (завершує поточну сесію).
func (s *AuthService) Logout(ctx context.Context, input auth.LogoutUserInput) (*auth.LogoutUserOutput, error) {
	return s.logoutUser.Execute(ctx, input)
}

// LogoutAll завершує всі сесії користувача.
func (s *AuthService) LogoutAll(ctx context.Context, input auth.LogoutAllUserSessionsInput) (*auth.LogoutUserOutput, error) {
	return s.logoutUser.ExecuteLogoutAll(ctx, input)
}

// ValidateSession перевіряє валідність сесії.
func (s *AuthService) ValidateSession(ctx context.Context, input auth.ValidateSessionInput) (*auth.ValidateSessionOutput, error) {
	return s.validateSession.Execute(ctx, input)
}

// ChangePassword змінює пароль користувача.
func (s *AuthService) ChangePassword(ctx context.Context, input auth.ChangePasswordInput) (*auth.ChangePasswordOutput, error) {
	return s.changePassword.Execute(ctx, input)
}

// CheckPermission перевіряє чи має користувач конкретний дозвіл.
func (s *AuthService) CheckPermission(ctx context.Context, input auth.CheckPermissionInput) (*auth.CheckPermissionOutput, error) {
	return s.checkPermission.Execute(ctx, input)
}

// CheckResourcePermission перевіряє чи має користувач дозвіл на дію над ресурсом.
func (s *AuthService) CheckResourcePermission(ctx context.Context, input auth.CheckResourcePermissionInput) (*auth.CheckPermissionOutput, error) {
	return s.checkPermission.ExecuteResourceCheck(ctx, input)
}

// RequirePermission викидає помилку якщо у користувача немає дозволу.
func (s *AuthService) RequirePermission(ctx context.Context, userID int64, permissionCode string) error {
	return s.checkPermission.RequirePermission(ctx, userID, permissionCode)
}

// RequireResourcePermission викидає помилку якщо у користувача немає дозволу на ресурс.
func (s *AuthService) RequireResourcePermission(ctx context.Context, userID int64, resource string, action entity.Action) error {
	return s.checkPermission.RequireResourcePermission(ctx, userID, resource, action)
}
