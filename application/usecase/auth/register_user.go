// application/usecase/auth/register_user.go
package auth

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
	"osbb-accounting/domain/service"
	"osbb-accounting/domain/valueobject"
)

// RegisterUserUseCase обробляє реєстрацію нового користувача.
type RegisterUserUseCase struct {
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	passwordHasher service.PasswordHasher
}

// NewRegisterUserUseCase створює новий use case.
func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordHasher service.PasswordHasher,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		passwordHasher: passwordHasher,
	}
}

// RegisterUserInput - вхідні дані для реєстрації.
type RegisterUserInput struct {
	Username   string
	Email      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName *string
	Phone      *string
	RoleName   string // Опціонально - роль за замовчуванням "viewer"
}

// RegisterUserOutput - результат реєстрації.
type RegisterUserOutput struct {
	UserID    int64
	Username  string
	Email     string
	FirstName string
	LastName  string
	Roles     []string
}

// Execute виконує реєстрацію користувача.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	// 1. Валідація вхідних даних через Value Object
	request, err := valueobject.NewRegistrationRequest(
		input.Username,
		input.Email,
		input.Password,
		input.FirstName,
		input.LastName,
		input.MiddleName,
		input.Phone,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid registration data",
			err,
		)
	}

	// 2. Валідація стійкості пароля
	if err := entity.ValidatePassword(request.Password); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"weak password",
			err,
		)
	}

	// 3. Перевірка унікальності username
	exists, err := uc.userRepo.ExistsByUsername(ctx, request.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"username already exists",
			domainErrors.ErrAlreadyExists,
		).WithDetails("field", "username")
	}

	// 4. Перевірка унікальності email
	exists, err = uc.userRepo.ExistsByEmail(ctx, request.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"email already exists",
			domainErrors.ErrAlreadyExists,
		).WithDetails("field", "email")
	}

	// 5. Хешування пароля
	passwordHash, err := uc.passwordHasher.Hash(request.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Створення Entity
	user, err := entity.NewUser(
		request.Username,
		request.Email,
		request.FirstName,
		request.LastName,
		request.MiddleName,
		request.Phone,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid user data",
			err,
		)
	}

	// Встановлюємо хеш пароля
	if err := user.SetPasswordHash(passwordHash); err != nil {
		return nil, fmt.Errorf("failed to set password hash: %w", err)
	}

	// 7. Збереження користувача
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 8. Призначення ролі за замовчуванням
	roleName := input.RoleName
	if roleName == "" {
		roleName = entity.RoleViewer // За замовчуванням - перегляд
	}

	role, err := uc.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		// Якщо роль не знайдена, призначаємо viewer
		role, err = uc.roleRepo.GetByName(ctx, entity.RoleViewer)
		if err != nil {
			return nil, fmt.Errorf("failed to get default role: %w", err)
		}
	}

	// Створюємо зв'язок користувач-роль
	userRole, err := entity.NewUserRole(user.ID, role.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user role: %w", err)
	}

	if err := uc.roleRepo.AssignToUser(ctx, userRole); err != nil {
		// Не критична помилка - користувач вже створений
		// В production тут краще використовувати транзакцію
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	// 9. Формування відповіді
	return &RegisterUserOutput{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     []string{role.Name},
	}, nil
}

// Validate перевіряє вхідні дані перед виконанням.
func (input *RegisterUserInput) Validate() error {
	validationErrors := domainErrors.NewValidationErrors()

	if input.Username == "" {
		validationErrors.Add("username", "username is required", input.Username)
	}
	if input.Email == "" {
		validationErrors.Add("email", "email is required", input.Email)
	}
	if input.Password == "" {
		validationErrors.Add("password", "password is required", input.Password)
	}
	if input.FirstName == "" {
		validationErrors.Add("first_name", "first name is required", input.FirstName)
	}
	if input.LastName == "" {
		validationErrors.Add("last_name", "last name is required", input.LastName)
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
