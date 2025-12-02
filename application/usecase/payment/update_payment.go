package payment

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
	"time"
)

// ============================================================================
// UpdatePaymentUseCase
// ============================================================================

type UpdatePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdatePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdatePaymentUseCase {
	return &UpdatePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdatePaymentInput struct {
	CurrentUserID  int64
	PaymentID      int64
	Amount         float64
	PaymentMethod  entity.PaymentMethod
	PaymentPurpose string
	PaymentDate    time.Time
	PeriodMonth    *int
	PeriodYear     *int
	ReceiptNumber  *string
	Notes          *string
}

func (uc *UpdatePaymentUseCase) Execute(ctx context.Context, input UpdatePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.IsApproved() {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"cannot update approved payment",
			nil,
		)
	}

	if err := payment.Update(
		input.Amount,
		input.PaymentMethod,
		input.PaymentPurpose,
		input.PaymentDate,
		input.ReceiptNumber,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}
