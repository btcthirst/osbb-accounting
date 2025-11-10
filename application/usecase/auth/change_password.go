// application/usecase/auth/change_password.go
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

// ChangePasswordUseCase обробляє зміну пароля користувача.
type ChangePasswordUseCase struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	passwordHasher service.PasswordHasher
}

// NewChangePasswordUseCase створює новий use case.
func NewChangePasswordUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	passwordHasher service.PasswordHasher,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		passwordHasher: passwordHasher,
	}
}

// ChangePasswordInput - вхідні дані для зміни пароля.
type ChangePasswordInput struct {
	UserID                  int64
	OldPassword             string
	NewPassword             string
	InvalidateOtherSessions bool // Чи завершити інші сесії після зміни пароля
}

// ChangePasswordOutput - результат зміни пароля.
type ChangePasswordOutput struct {
	Success             bool
	Message             string
	InvalidatedSessions int // Кількість завершених сесій
}

// Execute виконує зміну пароля.
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, input ChangePasswordInput) (*ChangePasswordOutput, error) {
	// 1. Валідація вхідних даних через Value Object
	request, err := valueobject.NewPasswordChangeRequest(
		input.UserID,
		input.OldPassword,
		input.NewPassword,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid password change request",
			err,
		)
	}

	// 2. Валідація стійкості нового пароля
	if err := entity.ValidatePassword(request.NewPassword); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"weak new password",
			err,
		)
	}

	// 3. Отримання користувача
	user, err := uc.userRepo.GetByID(ctx, request.UserID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 4. Перевірка старого пароля
	if err := uc.passwordHasher.Verify(request.OldPassword, user.PasswordHash); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidCredentials,
			"old password is incorrect",
			domainErrors.ErrInvalidCredentials,
		)
	}

	// 5. Хешування нового пароля
	newPasswordHash, err := uc.passwordHasher.Hash(request.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new password: %w", err)
	}

	// 6. Оновлення пароля
	if err := uc.userRepo.UpdatePasswordHash(ctx, user.ID, newPasswordHash); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// 7. Завершення інших сесій (опціонально)
	invalidatedSessions := 0
	if input.InvalidateOtherSessions {
		if err := uc.sessionRepo.DeleteAllByUserID(ctx, user.ID); err != nil {
			// Не критична помилка - пароль вже змінено
			// Можна залогувати для моніторингу
		} else {
			// Підрахувати кількість завершених сесій можна через GetActiveByUserID перед видаленням
			sessions, _ := uc.sessionRepo.GetActiveByUserID(ctx, user.ID)
			invalidatedSessions = len(sessions)
		}
	}

	return &ChangePasswordOutput{
		Success:             true,
		Message:             "password successfully changed",
		InvalidatedSessions: invalidatedSessions,
	}, nil
}

// Validate перевіряє вхідні дані перед виконанням.
func (input *ChangePasswordInput) Validate() error {
	validationErrors := domainErrors.NewValidationErrors()

	if input.UserID <= 0 {
		validationErrors.Add("user_id", "user ID is required", input.UserID)
	}
	if input.OldPassword == "" {
		validationErrors.Add("old_password", "old password is required", input.OldPassword)
	}
	if input.NewPassword == "" {
		validationErrors.Add("new_password", "new password is required", input.NewPassword)
	}
	if input.OldPassword == input.NewPassword {
		validationErrors.Add("new_password", "new password must be different from old password", input.NewPassword)
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
