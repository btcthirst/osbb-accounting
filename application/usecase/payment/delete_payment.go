package payment

import (
	"context"
	"fmt"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// DeletePaymentUseCase
// ============================================================================

type DeletePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewDeletePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *DeletePaymentUseCase {
	return &DeletePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type DeletePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *DeletePaymentUseCase) Execute(ctx context.Context, input DeletePaymentInput) (*shared.DeleteOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionDelete,
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
			"cannot delete approved payment",
			nil,
		)
	}

	if err := uc.paymentRepo.SoftDelete(ctx, input.PaymentID); err != nil {
		return nil, fmt.Errorf("failed to delete payment: %w", err)
	}

	return shared.NewDeleteOutput(
		"Payment successfully deleted",
	), nil
}
