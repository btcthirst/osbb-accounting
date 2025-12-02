package payment

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// ApprovePaymentUseCase
// ============================================================================

type ApprovePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewApprovePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ApprovePaymentUseCase {
	return &ApprovePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type ApprovePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *ApprovePaymentUseCase) Execute(ctx context.Context, input ApprovePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate, // Or a specific 'approve' action if defined
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

	if err := payment.Approve(input.CurrentUserID); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"failed to approve payment",
			err,
		)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// UnapprovePaymentUseCase
// ============================================================================

type UnapprovePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewUnapprovePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UnapprovePaymentUseCase {
	return &UnapprovePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type UnapprovePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *UnapprovePaymentUseCase) Execute(ctx context.Context, input UnapprovePaymentInput) (*PaymentOutput, error) {
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

	if err := payment.Unapprove(); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"failed to unapprove payment",
			err,
		)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}
