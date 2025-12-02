// application/usecase/charge/get_charge_statistics.go
package charge

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetChargeStatisticsUseCase отримує статистику нарахувань.
type GetChargeStatisticsUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

// NewGetChargeStatisticsUseCase створює новий use case.
func NewGetChargeStatisticsUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *GetChargeStatisticsUseCase {
	return &GetChargeStatisticsUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

// GetChargeStatisticsInput - вхідні дані.
type GetChargeStatisticsInput struct {
	CurrentUserID int64
}

// Execute виконує отримання статистики.
func (uc *GetChargeStatisticsUseCase) Execute(ctx context.Context, input GetChargeStatisticsInput) (*repository.ChargeStatistics, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Get statistics from repository
	stats, err := uc.chargeRepo.GetStatistics(ctx, repository.ChargeStatisticsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}
