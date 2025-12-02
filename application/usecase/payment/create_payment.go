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
// CreatePaymentUseCase
// ============================================================================

type CreatePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewCreatePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type CreatePaymentInput struct {
	CurrentUserID    int64
	OwnershipShareID int64
	Amount           float64
	PaymentMethod    entity.PaymentMethod
	PaymentPurpose   string
	PaymentDate      time.Time
	PeriodMonth      *int
	PeriodYear       *int
	ReceiptNumber    *string
	Notes            *string
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := entity.NewPayment(
		input.OwnershipShareID,
		input.Amount,
		input.PaymentMethod,
		input.PaymentPurpose,
		input.PaymentDate,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear
	payment.ReceiptNumber = input.ReceiptNumber
	payment.Notes = input.Notes

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}
