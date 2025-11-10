// application/usecase/auth/logout_user.go
package auth

import (
	"context"
	"fmt"

	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// LogoutUserUseCase обробляє вихід користувача з системи.
type LogoutUserUseCase struct {
	sessionRepo repository.SessionRepository
}

// NewLogoutUserUseCase створює новий use case.
func NewLogoutUserUseCase(sessionRepo repository.SessionRepository) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		sessionRepo: sessionRepo,
	}
}

// LogoutUserInput - вхідні дані для виходу.
type LogoutUserInput struct {
	SessionToken string
}

// LogoutUserOutput - результат виходу.
type LogoutUserOutput struct {
	Success bool
	Message string
}

// Execute виконує вихід користувача (видаляє сесію).
func (uc *LogoutUserUseCase) Execute(ctx context.Context, input LogoutUserInput) (*LogoutUserOutput, error) {
	// 1. Валідація токену
	if input.SessionToken == "" {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidToken,
			"session token is required",
			domainErrors.ErrInvalidToken,
		)
	}

	// 2. Видалення сесії
	err := uc.sessionRepo.DeleteByToken(ctx, input.SessionToken)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrSessionNotFound) {
			// Сесія вже не існує - це не помилка при logout
			return &LogoutUserOutput{
				Success: true,
				Message: "session already expired or logged out",
			}, nil
		}
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &LogoutUserOutput{
		Success: true,
		Message: "successfully logged out",
	}, nil
}

// LogoutAllUserSessions видаляє всі сесії користувача (logout з усіх пристроїв).
type LogoutAllUserSessionsInput struct {
	UserID int64
}

// ExecuteLogoutAll видаляє всі сесії користувача.
func (uc *LogoutUserUseCase) ExecuteLogoutAll(ctx context.Context, input LogoutAllUserSessionsInput) (*LogoutUserOutput, error) {
	if input.UserID <= 0 {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeInvalidInput,
			"invalid user ID",
			domainErrors.ErrInvalidInput,
		)
	}

	// Видаляємо всі сесії користувача
	if err := uc.sessionRepo.DeleteAllByUserID(ctx, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return &LogoutUserOutput{
		Success: true,
		Message: "successfully logged out from all devices",
	}, nil
}
