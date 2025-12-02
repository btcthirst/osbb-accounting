// application/usecase/apartment/get_apartment_statistics.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetApartmentStatisticsUseCase отримує статистику по квартирах.
type GetApartmentStatisticsUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewGetApartmentStatisticsUseCase створює новий use case.
func NewGetApartmentStatisticsUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *GetApartmentStatisticsUseCase {
	return &GetApartmentStatisticsUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// GetApartmentStatisticsInput - вхідні дані.
type GetApartmentStatisticsInput struct {
	CurrentUserID int64
}

// Execute виконує отримання статистики.
func (uc *GetApartmentStatisticsUseCase) Execute(ctx context.Context, input GetApartmentStatisticsInput) (*ApartmentStatisticsOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	stats, err := uc.apartmentRepo.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &ApartmentStatisticsOutput{
		TotalApartments: stats.TotalApartments,
		TotalArea:       stats.TotalArea,
		AverageArea:     stats.AverageArea,
		MinArea:         stats.MinArea,
		MaxArea:         stats.MaxArea,
		FloorCount:      stats.FloorCount,
		EntranceCount:   stats.EntranceCount,
		WithOwners:      stats.WithOwners,
		WithoutOwners:   stats.WithoutOwners,
	}, nil
}
