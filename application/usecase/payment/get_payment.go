package payment

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetPaymentUseCase
// ============================================================================

type GetPaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetPaymentUseCase {
	return &GetPaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, input GetPaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
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

	return mapPaymentToOutput(payment), nil
}
