package payment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// GetPaymentStatisticsUseCase
// ============================================================================

type GetPaymentStatisticsUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetPaymentStatisticsUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetPaymentStatisticsUseCase {
	return &GetPaymentStatisticsUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetPaymentStatisticsInput struct {
	CurrentUserID int64
	PeriodMonth   *int
	PeriodYear    *int
}

func (uc *GetPaymentStatisticsUseCase) Execute(ctx context.Context, input GetPaymentStatisticsInput) (*repository.PaymentStatistics, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.PaymentFilter{
		PeriodMonth: input.PeriodMonth,
		PeriodYear:  input.PeriodYear,
	}

	stats, err := uc.paymentRepo.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}
